package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pion/webrtc/v3"
	grpcpool "github.com/processout/grpc-go-pool"

	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/rtc"
	"github.com/aromancev/confa/internal/user/ident"
	"github.com/aromancev/confa/internal/user/session"
	"github.com/aromancev/confa/proto/queue"

	"github.com/aromancev/confa/internal/confa/talk"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/emails"
	"github.com/aromancev/confa/internal/platform/api"
	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/rtc/wsock"
	"github.com/aromancev/confa/proto/sfu"
)

type accessToken struct {
	Token     string `json:"token"`
	ExpiresIn uint64 `json:"expiresIn"`
}

func ok(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	_, _ = w.Write([]byte("OK"))
}

type loginReq struct {
	Email string `json:"email"`
}

func (r loginReq) Validate() error {
	if err := email.Validate(r.Email); err != nil {
		return err
	}
	return nil
}

func login(baseURL string, signer *auth.Signer, producer Producer) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()

		var req loginReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = api.BadRequest(w, api.CodeMalformedRequest, err.Error())
			return
		}

		if err := req.Validate(); err != nil {
			_ = api.BadRequest(w, api.CodeInvalidRequest, err.Error())
			return
		}

		token, err := signer.EmailToken(req.Email)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to create email token")
			_ = api.InternalError(w)
			return
		}

		msg, err := emails.Login(baseURL, req.Email, token)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to render login email")
			_ = api.InternalError(w)
			return
		}

		body, err := queue.Marshal(&queue.EmailJob{
			Emails: []*queue.Email{{
				FromName:  msg.FromName,
				ToAddress: msg.ToAddress,
				Subject:   msg.Subject,
				Html:      msg.HTML,
			}},
		}, trace.ID(ctx))
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to marshal email")
			_ = api.InternalError(w)
			return
		}

		id, err := producer.Put(ctx, queue.TubeEmail, body, beanstalk.PutParams{})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to put email job")
			_ = api.InternalError(w)
			return
		}
		log.Ctx(ctx).Info().Uint64("jobId", id).Msg("Email login job emitted")
	}
}

func createSession(verifier *auth.Verifier, signer *auth.Signer, idents *ident.CRUD, sessions *session.CRUD) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()

		claims, err := verifier.EmailToken(auth.Bearer(r))
		if err != nil {
			_ = api.Unauthorised(w)
			return
		}

		userID, err := idents.GetOrCreate(ctx, ident.Ident{
			Platform: ident.PlatformEmail,
			Value:    claims.Address,
		})
		if err != nil {
			_ = api.Unauthorised(w)
			return
		}

		sess, err := sessions.Create(ctx, userID)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to create session")
			_ = api.InternalError(w)
			return
		}

		access, expiresIn, err := signer.AccessToken(userID)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to sign")
			_ = api.InternalError(w)
			return
		}

		auth.SetSession(w, sess.Key)
		_ = api.Created(w, accessToken{
			Token:     access,
			ExpiresIn: uint64(expiresIn.Seconds()),
		})
	}
}

func createToken(signer *auth.Signer, sessions *session.CRUD) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()

		sess, err := sessions.Fetch(ctx, auth.Session(r))
		if err != nil {
			_ = api.Unauthorised(w)
			return
		}

		access, expiresIn, err := signer.AccessToken(sess.Owner)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to sign")
			_ = api.InternalError(w)
			return
		}

		_ = api.Created(w, accessToken{
			Token:     access,
			ExpiresIn: uint64(expiresIn.Seconds()),
		})
	}
}

func createConfa(verifier *auth.Verifier, confas *confa.CRUD) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()

		access, err := verifier.AccessToken(auth.Bearer(r))
		if err != nil {
			_ = api.Unauthorised(w)
			return
		}

		var request confa.Confa
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			_ = api.BadRequest(w, api.CodeMalformedRequest, err.Error())
			return
		}

		conf, err := confas.Create(ctx, access.UserID, request)
		switch {
		case errors.Is(err, confa.ErrValidation):
			_ = api.BadRequest(w, api.CodeInvalidRequest, err.Error())
			return
		case errors.Is(err, confa.ErrDuplicatedEntry):
			_ = api.BadRequest(w, api.CodeDuplicatedEntry, err.Error())
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to create confa")
			_ = api.InternalError(w)
			return
		}

		_ = api.Created(w, conf)
	}
}

func getConfa(confas *confa.CRUD) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()

		confID, err := uuid.Parse(ps.ByName("confa_id"))
		if err != nil {
			_ = api.NotFound(w, err.Error())
			return
		}
		conf, err := confas.Fetch(ctx, confID)
		switch {
		case errors.Is(err, confa.ErrNotFound):
			_ = api.NotFound(w, err.Error())
			return
		case errors.Is(err, confa.ErrValidation):
			_ = api.BadRequest(w, api.CodeInvalidRequest, err.Error())
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to fetch confa")
			_ = api.InternalError(w)
			return
		}

		_ = api.OK(w, conf)
	}
}

func createTalk(verifier *auth.Verifier, talks *talk.CRUD) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()

		access, err := verifier.AccessToken(auth.Bearer(r))
		if err != nil {
			_ = api.Unauthorised(w)
			return
		}

		var request talk.Talk
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			_ = api.BadRequest(w, api.CodeMalformedRequest, err.Error())
			return
		}

		confaID, err := uuid.Parse(ps.ByName("confa_id"))
		if err != nil {
			_ = api.NotFound(w, err.Error())
			return
		}

		tlk, err := talks.Create(ctx, confaID, access.UserID, request)
		switch {
		case errors.Is(err, talk.ErrValidation):
			_ = api.BadRequest(w, api.CodeInvalidRequest, err.Error())
			return
		case errors.Is(err, talk.ErrDuplicatedEntry):
			_ = api.BadRequest(w, api.CodeDuplicatedEntry, err.Error())
			return
		case errors.Is(err, talk.ErrPermissionDenied):
			_ = api.Forbidden(w)
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to create talk")
			_ = api.InternalError(w)
			return
		}

		_ = api.Created(w, tlk)
	}
}

func getTalk(talks *talk.CRUD) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()

		talkID, err := uuid.Parse(ps.ByName("talk_id"))
		if err != nil {
			_ = api.NotFound(w, err.Error())
			return
		}

		tlk, err := talks.Fetch(ctx, talkID)
		switch {
		case errors.Is(err, talk.ErrNotFound):
			_ = api.NotFound(w, err.Error())
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to fetch talk")
			_ = api.InternalError(w)
			return
		}

		_ = api.OK(w, tlk)
	}
}

func serveRTC(upgrader *wsock.Upgrader, sfuPool, mediaPool *grpcpool.Pool) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()

		conn, err := upgrader.NewConn(w, r, nil)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to upgrade connection.")
			return
		}
		defer conn.Close()

		peer, err := sfu.NewPeer(ctx, sfuPool)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to create new SFU peer.")
			return
		}
		defer peer.Close()

		peer.OnOffer(func(offer webrtc.SessionDescription) {
			err := conn.NotifyOffer(offer)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to notify about offer.")
				return
			}
		})
		peer.OnTrickle(func(target int, candidate webrtc.ICECandidateInit) {
			err := conn.NotifyTrickle(target, candidate)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to notify about trickle.")
				return
			}
		})

		sess := rtc.NewSession(mediaPool, peer)
		for {
			request, err := conn.Receive()
			if err != nil {
				if !errors.Is(err, wsock.ErrClosed) {
					log.Ctx(ctx).Err(err).Msg("Failed to receive message.")
				}
				return
			}

			switch req := request.(type) {
			case wsock.Join:
				answer, err := sess.Join(ctx, req.SID, req.UID, req.Offer)
				if err != nil {
					log.Ctx(ctx).Err(err).Msg("Failed to join.")
					_ = req.Error(err.Error())
					return
				}
				_ = req.Reply(answer)

			case wsock.Offer:
				answer, err := sess.Offer(ctx, req.Offer)
				switch {
				case errors.Is(err, rtc.ErrValidation):
					_ = req.Error(err.Error())
					return
				case err != nil:
					log.Ctx(ctx).Err(err).Msg("Failed to receive offer.")
					_ = req.Error(err.Error())
					return
				}
				_ = req.Reply(answer)

			case wsock.Answer:
				err = peer.Answer(ctx, req.Answer)
				if err != nil {
					log.Ctx(ctx).Err(err).Msg("Failed to receive answer.")
					_ = req.Error(err.Error())
				}

			case wsock.Trickle:
				err = peer.Trickle(ctx, req.Target, req.Candidate)
				if err != nil {
					log.Ctx(ctx).Err(err).Msg("Failed to receive trickle.")
					_ = req.Error(err.Error())
				}
			}
		}
	}
}

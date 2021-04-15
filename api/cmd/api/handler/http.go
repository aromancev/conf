package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	rpcws "github.com/sourcegraph/jsonrpc2/websocket"

	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/user/ident"
	"github.com/aromancev/confa/internal/user/session"

	"github.com/aromancev/confa/internal/confa/talk"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/emails"
	"github.com/aromancev/confa/internal/platform/api"
	"github.com/aromancev/confa/internal/platform/email"
)

type accessToken struct {
	Token     string `json:"token"`
	ExpiresIn uint64 `json:"expiresIn"`
}

type loginReq struct {
	Email string `json:"email"`
}

func (r loginReq) Validate() error {
	if err := email.ValidateEmail(r.Email); err != nil {
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
		body, err := json.Marshal([]email.Email{msg})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to marshal email")
			_ = api.InternalError(w)
			return
		}

		id, err := producer.Put(ctx, TubeEmail, body, beanstalk.PutParams{})
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

func ok(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	_, _ = w.Write([]byte("OK"))
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

func rtc(upgrader websocket.Upgrader, sfuAddress string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to upgrade connection.")
			return
		}
		defer c.Close()

		signal, err := NewSignal(ctx, sfuAddress)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to start new signal.")
			return
		}
		defer signal.Close()

		conn := jsonrpc2.NewConn(ctx, rpcws.NewObjectStream(c), signal)
		go func() {
			if err := signal.Serve(ctx, conn); err != nil {
				if errors.Is(err, context.Canceled) {
					log.Ctx(ctx).Err(err).Msg("Failed to serve signal.")
				}
			}
		}()

		<-conn.DisconnectNotify()
		log.Ctx(ctx).Info().Msg("RTC client disconnected.")
	}
}

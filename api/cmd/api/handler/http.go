package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pion/webrtc/v3"

	"github.com/aromancev/confa/internal/platform/sfu"
	"github.com/aromancev/confa/internal/rtc"

	"github.com/aromancev/confa/internal/confa/talk"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/platform/api"
	"github.com/aromancev/confa/internal/rtc/wsock"
)

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

func serveRTC(upgrader *wsock.Upgrader, sfuAddr string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()

		conn, err := upgrader.NewConn(w, r, nil)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to upgrade connection.")
			return
		}
		defer conn.Close()

		signal, err := sfu.NewSignal(ctx, sfuAddr)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to create new signal.")
			return
		}
		defer signal.Close()

		sess := rtc.NewSession(signal)
		sess.OnAnswer(func(desc webrtc.SessionDescription) {
			err := conn.Answer(desc)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to notify about answer.")
			}
		})
		sess.OnOffer(func(desc webrtc.SessionDescription) {
			err := conn.Offer(desc)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to notify about offer.")
			}
		})
		sess.OnTrickle(func(cand webrtc.ICECandidateInit, target int) {
			err := conn.Trickle(cand, target)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to notify about trickle.")
			}
		})

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
				err := sess.Join(ctx, req.SID, req.UID, req.Offer)
				if err != nil {
					log.Ctx(ctx).Err(err).Msg("Failed to join.")
					_ = conn.Error(err.Error())
					return
				}

			case wsock.Offer:
				err := sess.Offer(ctx, req.Description)
				switch {
				case errors.Is(err, rtc.ErrValidation):
					_ = conn.Error(err.Error())
					return
				case err != nil:
					log.Ctx(ctx).Err(err).Msg("Failed to receive offer.")
					_ = conn.Error(err.Error())
					return
				}

			case wsock.Answer:
				err := sess.Answer(ctx, req.Description)
				if err != nil {
					log.Ctx(ctx).Err(err).Msg("Failed to answer.")
					_ = conn.Error(err.Error())
					return
				}

			case wsock.Trickle:
				err := sess.Trickle(ctx, req.Candidate, req.Target)
				if err != nil {
					log.Ctx(ctx).Err(err).Msg("Failed to trickle.")
					_ = conn.Error(err.Error())
					return
				}
			}
		}
	}
}

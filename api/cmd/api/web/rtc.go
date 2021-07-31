package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/internal/platform/sfu"
	"github.com/aromancev/confa/internal/room"
	"github.com/aromancev/confa/internal/room/peer"
	"github.com/aromancev/confa/internal/room/peer/wsock"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
)

func serveRTC(rooms *room.Mongo, pk *auth.PublicKey, upgrader *wsock.Upgrader, sfuAddr string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var claims auth.APIClaims
		if err := pk.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 4 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		roomID, err := uuid.Parse(parts[3])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		rm, err := rooms.FetchOne(ctx, room.Lookup{ID: roomID})
		switch {
		case errors.Is(err, room.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
			return
		case err != nil:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

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

		p := peer.NewPeer(rm, signal)
		p.OnAnswer(func(desc webrtc.SessionDescription) {
			err := conn.Answer(desc)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to notify about answer.")
			}
		})
		p.OnOffer(func(desc webrtc.SessionDescription) {
			err := conn.Offer(desc)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to notify about offer.")
			}
		})
		p.OnTrickle(func(cand webrtc.ICECandidateInit, target int) {
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
				err := p.Join(ctx, req.SID, req.UID, req.Offer)
				if err != nil {
					log.Ctx(ctx).Err(err).Msg("Failed to join.")
					_ = conn.Error(err.Error())
					return
				}

			case wsock.Offer:
				err := p.Offer(ctx, req.Description)
				switch {
				case errors.Is(err, peer.ErrValidation):
					_ = conn.Error(err.Error())
					return
				case err != nil:
					log.Ctx(ctx).Err(err).Msg("Failed to receive offer.")
					_ = conn.Error(err.Error())
					return
				}

			case wsock.Answer:
				err := p.Answer(ctx, req.Description)
				if err != nil {
					log.Ctx(ctx).Err(err).Msg("Failed to answer.")
					_ = conn.Error(err.Error())
					return
				}

			case wsock.Trickle:
				err := p.Trickle(ctx, req.Candidate, req.Target)
				if err != nil {
					log.Ctx(ctx).Err(err).Msg("Failed to trickle.")
					_ = conn.Error(err.Error())
					return
				}
			}
		}
	}
}

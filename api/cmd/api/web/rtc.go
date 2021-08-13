package web

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/internal/event/peer"
	"github.com/aromancev/confa/internal/platform/grpcpool"
	"github.com/aromancev/confa/internal/room"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

func serveRTC(rooms *room.Mongo, pk *auth.PublicKey, upgrader *websocket.Upgrader, sfuPool *grpcpool.Pool) func(http.ResponseWriter, *http.Request) {
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

		_, err = rooms.FetchOne(ctx, room.Lookup{ID: roomID})
		switch {
		case errors.Is(err, room.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
			return
		case err != nil:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to upgrade websocket connection.")
			return
		}
		defer wsConn.Close()
		log.Ctx(ctx).Info().Msg("Websocket connected.")

		peerConn, err := peer.NewPeer(ctx, sfuPool)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to connect to peer.")
			return
		}
		defer peerConn.Close(ctx)
		log.Ctx(ctx).Info().Msg("Peer connected.")

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()

			for {
				msg, err := peerConn.Receive(ctx)
				switch {
				case errors.Is(err, peer.ErrClosed):
					_ = wsConn.Close()
					log.Ctx(ctx).Info().Msg("Peer disconnected.")
					return
				case errors.Is(err, peer.ErrUnknownMessage):
					log.Ctx(ctx).Debug().Msg("Skipping unknown message from peer.")
				case err != nil:
					log.Ctx(ctx).Err(err).Msg("Failed to receive message from peer.")
					return
				}

				err = wsConn.WriteJSON(msg)
				if err != nil {
					log.Ctx(ctx).Err(err).Msg("Failed send message to websocket.")
				}
			}
		}()

		go func() {
			defer wg.Done()

			for {
				var msg peer.Message
				err := wsConn.ReadJSON(&msg)
				switch {
				case websocket.IsCloseError(err, websocket.CloseGoingAway):
					_ = peerConn.Close(ctx)
					log.Ctx(ctx).Info().Msg("Websocket disconnected.")
					return
				case err != nil:
					log.Ctx(ctx).Err(err).Msg("Failed to receive message from websocket.")
					return
				}

				err = peerConn.Send(ctx, msg)
				switch {
				case errors.Is(err, peer.ErrValidation):
				case err != nil:
					log.Ctx(ctx).Err(err).Msg("Failed send peer message.")
				}
			}
		}()

		wg.Wait()
	}
}

package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/event/peer"
	"github.com/aromancev/confa/event/peer/signal"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/room"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Handler struct {
	router http.Handler
}

func NewHandler(resolver *Resolver, pk *auth.PublicKey, rooms *room.Mongo, upgrader *websocket.Upgrader, producer Producer, sfuConn *grpc.ClientConn, eventWatcher event.Watcher) *Handler {
	r := http.NewServeMux()

	r.HandleFunc("/health", ok)
	r.Handle(
		"/query",
		withHTTPAuth(
			&relay.Handler{
				Schema: graphql.MustParseSchema(schema, resolver, graphql.UseFieldResolvers()),
			},
		),
	)
	r.Handle(
		"/room/",
		withNewTrace(
			withWSockAuth(
				serveRTC(rooms, pk, upgrader, sfuConn, producer, eventWatcher),
			),
		),
	)

	return &Handler{
		router: r,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	traceID := r.Header.Get("Trace-Id")
	if traceID != "" {
		ctx = trace.New(ctx, traceID)
	}

	defer func() {
		if err := recover(); err != nil {
			log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Msg("ServeHTTP panic")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	h.router.ServeHTTP(w, r.WithContext(ctx))
}

func ok(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func serveRTC(rooms *room.Mongo, pk *auth.PublicKey, upgrader *websocket.Upgrader, sfuConn *grpc.ClientConn, producer Producer, events event.Watcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		var claims auth.APIClaims
		if err := pk.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 3 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		roomID, err := uuid.Parse(parts[2])
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

		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to upgrade websocket connection.")
			return
		}
		defer wsConn.Close()
		log.Ctx(ctx).Debug().Msg("Websocket connected.")

		cursor, err := events.Watch(ctx, rm.ID)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to connect watch room events.")
			return
		}
		log.Ctx(ctx).Debug().Msg("Event watching started.")

		peerConn := peer.NewPeer(ctx, claims.UserID, rm.ID, signal.NewGRPCSignal(ctx, sfuConn), cursor, producer, 10)
		defer peerConn.Close(ctx)
		log.Ctx(ctx).Debug().Msg("Peer connected.")

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()

			for {
				msg, err := peerConn.Receive(ctx)
				switch {
				case errors.Is(err, peer.ErrClosed), errors.Is(err, context.Canceled):
					cancel() // If peer closed for any reason, terminate the whole connection.
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
					log.Ctx(ctx).Warn().Err(err).Msg("Failed send message to websocket. Closing.")
					cancel()
				}
			}
		}()

		go func() {
			defer wg.Done()

			for {
				var msg peer.Message
				err := wsConn.ReadJSON(&msg)
				switch {
				case websocket.IsCloseError(err, websocket.CloseGoingAway), errors.Is(err, context.Canceled):
					cancel() // If ws closed for any reason, terminate the whole connection.
					log.Ctx(ctx).Info().Msg("Websocket disconnected.")
					return
				case err != nil:
					log.Ctx(ctx).Warn().Err(err).Msg("Failed to receive message from websocket.")
					return
				}

				err = peerConn.Send(ctx, msg)
				switch {
				case errors.Is(err, peer.ErrValidation):
					log.Ctx(ctx).Warn().Err(err).Msg("Message from websocket rejected.")
				case errors.Is(err, context.Canceled):
					cancel() // If peer closed for any reason, terminate the whole connection.
					log.Ctx(ctx).Info().Msg("Peer disconnected.")
					return
				case err != nil:
					log.Ctx(ctx).Err(err).Msg("Failed send peer message.")
				}
			}
		}()

		log.Ctx(ctx).Info().Str("roomId", rm.ID.String()).Msg("RTC peer connected.")
		wg.Wait()
	})
}

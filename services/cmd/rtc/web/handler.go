package web

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/event/proxy"
	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/room"
	"github.com/aromancev/confa/room/record"
	"github.com/google/uuid"
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

func NewHandler(pk *auth.PublicKey, rooms *room.Mongo, events *event.Mongo, emitter proxy.EventEmitter, sfuConn *grpc.ClientConn, eventWatcher event.Watcher, recs *record.Mongo) *Handler {
	r := http.NewServeMux()

	r.HandleFunc("/health", ok)
	r.Handle(
		"/graph",
		withHTTPAuth(
			&relay.Handler{
				Schema: graphql.MustParseSchema(
					gqlSchema,
					NewResolver(pk, events, recs),
					graphql.UseFieldResolvers(),
				),
			},
		),
	)
	r.HandleFunc(
		"/room/schema",
		serveRoomSchema,
	)
	r.Handle(
		"/room/socket/",
		withNewTrace(
			withWebSocketAuth(
				roomWebSocket(rooms, pk, sfuConn, emitter, eventWatcher),
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
			log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Bytes("stack", debug.Stack()).Msg("ServeHTTP panic")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	h.router.ServeHTTP(w, r.WithContext(ctx))
}

func roomWebSocket(rooms *room.Mongo, pk *auth.PublicKey, sfuConn *grpc.ClientConn, emitter proxy.EventEmitter, events event.Watcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		defer func() {
			if err := recover(); err != nil {
				log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Bytes("stack", debug.Stack()).Msg("WebSocket handler panic")
			}
		}()

		var claims auth.APIClaims
		if err := pk.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 4 {
			log.Ctx(ctx).Debug().Str("url", r.URL.Path).Msg("Unexpected URL pattern")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		roomID, err := uuid.Parse(parts[3])
		if err != nil {
			log.Ctx(ctx).Debug().Str("url", r.URL.Path).Msg("Unexpected URL pattern")
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

		pp, err := NewPeerProxy(ctx, w, r, claims.UserID, rm.ID, events, emitter, sfuConn)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to connect to websocket.")
			return
		}
		defer pp.Close(ctx)

		log.Ctx(ctx).Info().Str("roomId", rm.ID.String()).Str("userId", claims.UserID.String()).Str("sessionId", pp.SessionID().String()).Msg("RTC peer connected.")
		pp.Serve()
		log.Ctx(ctx).Info().Str("roomId", rm.ID.String()).Str("userId", claims.UserID.String()).Str("sessionId", pp.SessionID().String()).Msg("RTC peer disconnected.")
	})
}

func ok(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func serveRoomSchema(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(roomSchema))
}

//go:embed room.schema.json
var roomSchema string

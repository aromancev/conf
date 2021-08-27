package web

import (
	"context"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/aromancev/confa/internal/confa/talk/clap"
	"github.com/aromancev/confa/internal/event"
	"github.com/aromancev/confa/internal/platform/grpcpool"
	"github.com/aromancev/confa/internal/room"
	"github.com/aromancev/confa/internal/user"
	"github.com/aromancev/confa/internal/user/session"
	"github.com/gorilla/websocket"
)

type EventRepo interface {
	Fetch(ctx context.Context, lookup event.Lookup) ([]event.Event, error)
}

type Resolver struct {
	baseURL      string
	secretKey    *auth.SecretKey
	publicKey    *auth.PublicKey
	users        *user.CRUD
	sessions     *session.CRUD
	confas       *confa.CRUD
	talks        *talk.CRUD
	claps        *clap.CRUD
	rooms        *room.Mongo
	events       EventRepo
	eventWatcher event.Watcher
	producer     Producer
	upgrader     *websocket.Upgrader
	sfuPool      *grpcpool.Pool
}

func NewResolver(baseURL string, sk *auth.SecretKey, pk *auth.PublicKey, producer Producer, users *user.CRUD, sessions *session.CRUD, confas *confa.CRUD, talks *talk.CRUD, claps *clap.CRUD, rooms *room.Mongo, upgrader *websocket.Upgrader, sfuPool *grpcpool.Pool, eventWatcher event.Watcher, events EventRepo) *Resolver {
	return &Resolver{
		baseURL:      baseURL,
		secretKey:    sk,
		publicKey:    pk,
		producer:     producer,
		users:        users,
		sessions:     sessions,
		confas:       confas,
		talks:        talks,
		claps:        claps,
		rooms:        rooms,
		upgrader:     upgrader,
		sfuPool:      sfuPool,
		eventWatcher: eventWatcher,
		events:       events,
	}
}

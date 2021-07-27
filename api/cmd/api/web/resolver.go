package web

import (
	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/aromancev/confa/internal/confa/talk/clap"
	"github.com/aromancev/confa/internal/room"
	"github.com/aromancev/confa/internal/room/peer/wsock"
	"github.com/aromancev/confa/internal/user"
	"github.com/aromancev/confa/internal/user/session"
)

type Resolver struct {
	baseURL   string
	secretKey *auth.SecretKey
	publicKey *auth.PublicKey
	users     *user.CRUD
	sessions  *session.CRUD
	confas    *confa.CRUD
	talks     *talk.CRUD
	claps     *clap.CRUD
	rooms     *room.Mongo
	producer  Producer
	upgrader  *wsock.Upgrader
	sfuAddr   string
}

func NewResolver(baseURL string, sk *auth.SecretKey, pk *auth.PublicKey, producer Producer, users *user.CRUD, sessions *session.CRUD, confas *confa.CRUD, talks *talk.CRUD, claps *clap.CRUD, rooms *room.Mongo, upgrader *wsock.Upgrader, sfuAddr string) *Resolver {
	return &Resolver{
		baseURL:   baseURL,
		secretKey: sk,
		publicKey: pk,
		producer:  producer,
		users:     users,
		sessions:  sessions,
		confas:    confas,
		talks:     talks,
		claps:     claps,
		rooms:     rooms,
		upgrader:  upgrader,
		sfuAddr:   sfuAddr,
	}
}

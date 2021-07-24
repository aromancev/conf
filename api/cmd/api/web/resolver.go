package web

import (
	"context"

	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/aromancev/confa/internal/confa/talk/clap"
	"github.com/aromancev/confa/internal/rtc/wsock"
	"github.com/aromancev/confa/internal/user"
	"github.com/aromancev/confa/internal/user/session"
	"github.com/prep/beanstalk"
)

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Resolver struct {
	baseURL   string
	secretKey *auth.SecretKey
	publicKey *auth.PublicKey
	users     *user.CRUD
	sessions  *session.CRUD
	confas    *confa.CRUD
	talks     *talk.CRUD
	claps     *clap.CRUD
	producer  Producer
	upgrader  *wsock.Upgrader
	sfuAddr   string
}

func NewResolver(baseURL string, sk *auth.SecretKey, pk *auth.PublicKey, producer Producer, users *user.CRUD, sessions *session.CRUD, confas *confa.CRUD, talks *talk.CRUD, claps *clap.CRUD, upgrader *wsock.Upgrader, sfuAddr string) *Resolver {
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
		upgrader:  upgrader,
		sfuAddr:   sfuAddr,
	}
}

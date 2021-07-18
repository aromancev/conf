package web

import (
	"context"
	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/aromancev/confa/internal/confa/talk/clap"

	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/user/ident"
	"github.com/aromancev/confa/internal/user/session"
	"github.com/prep/beanstalk"
)

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Resolver struct {
	baseURL  string
	signer   *auth.Signer
	verifier *auth.Verifier
	idents   *ident.CRUD
	sessions *session.CRUD
	confas   *confa.CRUD
	talks    *talk.CRUD
	claps    *clap.CRUD
	producer Producer
}

func NewResolver(baseURL string, signer *auth.Signer, verifier *auth.Verifier, producer Producer, idents *ident.CRUD, sessions *session.CRUD, confas *confa.CRUD, talks *talk.CRUD, claps *clap.CRUD) *Resolver {
	return &Resolver{
		baseURL:  baseURL,
		signer:   signer,
		verifier: verifier,
		producer: producer,
		idents:   idents,
		sessions: sessions,
		confas:   confas,
		talks:    talks,
		claps:    claps,
	}
}

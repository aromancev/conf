package talk

import (
	"context"
	"fmt"

	"github.com/aromancev/confa/internal/confa"

	"github.com/google/uuid"

	"github.com/aromancev/confa/internal/platform/psql"
)

type Repo interface {
	Create(ctx context.Context, execer psql.Execer, requests ...Talk) ([]Talk, error)
	Fetch(ctx context.Context, queryer psql.Queryer, lookup Lookup) ([]Talk, error)
	FetchOne(ctx context.Context, queryer psql.Queryer, lookup Lookup) (Talk, error)
}

type ConfaRepo interface {
	FetchOne(ctx context.Context, lookup confa.Lookup) (confa.Confa, error)
}

type CRUD struct {
	conn   psql.Conn
	repo   Repo
	confas ConfaRepo
}

func NewCRUD(conn psql.Conn, repo Repo, confas ConfaRepo) *CRUD {
	return &CRUD{conn: conn, repo: repo, confas: confas}
}

func (c *CRUD) Create(ctx context.Context, confaID, ownerID uuid.UUID, request Talk) (Talk, error) {
	request.ID = uuid.New()
	request.Confa = confaID
	request.Owner = ownerID
	request.Speaker = ownerID
	if request.Handle == "" {
		request.Handle = request.ID.String()
	}
	if err := request.Validate(); err != nil {
		return Talk{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	conf, err := c.confas.FetchOne(ctx, confa.Lookup{ID: confaID})
	if err != nil {
		return Talk{}, fmt.Errorf("failed to fetch confa: %w", err)
	}

	if conf.Owner != ownerID {
		return Talk{}, ErrPermissionDenied
	}

	created, err := c.repo.Create(ctx, c.conn, request)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to create talk: %w", err)
	}

	return created[0], nil
}

func (c *CRUD) Fetch(ctx context.Context, lookup Lookup) ([]Talk, error) {
	fetched, err := c.repo.Fetch(ctx, c.conn, lookup)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch talk: %w", err)
	}
	return fetched, nil
}

func (c *CRUD) FetchOne(ctx context.Context, lookup Lookup) (Talk, error) {
	fetched, err := c.repo.FetchOne(ctx, c.conn, lookup)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to fetch talk: %w", err)
	}
	return fetched, nil
}

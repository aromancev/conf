package talk

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/aromancev/confa/internal/platform/psql"
)

type Repo interface {
	Create(ctx context.Context, execer psql.Execer, requests ...Talk) ([]Talk, error)
	FetchOne(ctx context.Context, queryer psql.Queryer, lookup Lookup) (Talk, error)
}

type CRUD struct {
	conn psql.Conn
	repo Repo
}

func NewCRUD(conn psql.Conn, repo Repo) *CRUD {
	return &CRUD{conn: conn, repo: repo}
}

func (c *CRUD) Create(ctx context.Context, confaID uuid.UUID, request Talk) (Talk, error) {
	request.ID = uuid.New()
	request.Confa = confaID
	if err := request.Validate(); err != nil {
		return Talk{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}

	created, err := c.repo.Create(ctx, c.conn, request)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to create confa: %w", err)
	}

	return created[0], nil
}

func (c *CRUD) Fetch(ctx context.Context, ID uuid.UUID) (Talk, error) {

	fetched, err := c.repo.FetchOne(ctx, c.conn, Lookup{ID: ID})
	if err != nil {
		return Talk{}, fmt.Errorf("failed to fetch confa: %w", err)
	}
	return fetched, nil
}
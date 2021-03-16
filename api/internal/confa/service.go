package confa

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/aromancev/confa/internal/platform/psql"
)

type Repo interface {
	Create(ctx context.Context, execer psql.Execer, requests ...Confa) ([]Confa, error)
	FetchOne(ctx context.Context, queryer psql.Queryer, lookup Lookup) (Confa, error)
}

type CRUD struct {
	conn psql.Conn
	repo Repo
}

func NewCRUD(conn psql.Conn, repo Repo) *CRUD {
	return &CRUD{conn: conn, repo: repo}
}

func (c *CRUD) Create(ctx context.Context, userID uuid.UUID, request Confa) (Confa, error) {
	request.ID = uuid.New()
	request.Owner = userID
	if err := request.Validate(); err != nil {
		return Confa{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}

	created, err := c.repo.Create(ctx, c.conn, request)
	if err != nil {
		return Confa{}, fmt.Errorf("failed to create confa: %w", err)
	}

	return created[0], nil
}

func (c *CRUD) Fetch(ctx context.Context, ID uuid.UUID) (Confa, error) {

	fetched, err := c.repo.FetchOne(ctx, c.conn, Lookup{ID: ID})
	if err != nil {
		return Confa{}, fmt.Errorf("failed to fetch confa: %w", err)
	}
	return fetched, nil
}

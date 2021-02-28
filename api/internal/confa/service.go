package confa

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/aromancev/confa/internal/platform/psql"
)

type Repo interface {
	Create(ctx context.Context, execer psql.Execer, requests ...Confa) ([]Confa, error)
	Get(ctx context.Context, queryer psql.Queryer, ids ...uuid.UUID) ([]Confa, error)
}

type CRUD struct {
	db   *sql.DB
	repo Repo
}

func NewCRUD(db *sql.DB, repo Repo) *CRUD {
	return &CRUD{db: db, repo: repo}
}

func (c *CRUD) Create(ctx context.Context, userID uuid.UUID, request Confa) (Confa, error) {
	request.ID = uuid.New()
	request.Owner = userID
	if err := request.Validate(); err != nil {
		return Confa{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}

	created, err := c.repo.Create(ctx, c.db, request)
	if err != nil {
		return Confa{}, fmt.Errorf("failed to create confa: %w", err)
	}

	return created[0], nil
}

func (c *CRUD) Confa(ctx context.Context, userID uuid.UUID, id uuid.UUID) (Confa, error) {
	created, err := c.repo.Get(ctx, c.db, id)
	if err != nil {
		return Confa{}, fmt.Errorf("failed to create confa: %w", err)
	}

	return created[0], nil
}
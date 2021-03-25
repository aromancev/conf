package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/aromancev/confa/internal/platform/psql"
)

type Repo interface {
	Create(ctx context.Context, execer psql.Execer, requests ...User) ([]User, error)
	FetchOne(ctx context.Context, queryer psql.Queryer, lookup Lookup) (User, error)
}

type CRUD struct {
	conn psql.Conn
	repo Repo
}

func NewCRUD(conn psql.Conn, repo Repo) *CRUD {
	return &CRUD{conn: conn, repo: repo}
}

func (c *CRUD) Create(ctx context.Context, request User) (User, error) {
	request.ID = uuid.New()
	if err := request.Validate(); err != nil {
		return User{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}

	created, err := c.repo.Create(ctx, c.conn, request)
	if err != nil {
		return User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return created[0], nil
}

func (c *CRUD) Fetch(ctx context.Context, ID uuid.UUID) (User, error) {
	fetched, err := c.repo.FetchOne(ctx, c.conn, Lookup{ID: ID})
	if err != nil {
		return User{}, fmt.Errorf("failed to fetch user: %w", err)
	}
	return fetched, nil
}

package session

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/aromancev/confa/internal/platform/psql"
)

type Repo interface {
	Create(ctx context.Context, execer psql.Execer, requests ...Session) ([]Session, error)
	FetchOne(ctx context.Context, queryer psql.Queryer, lookup Lookup) (Session, error)
}

type CRUD struct {
	conn psql.Conn
	repo Repo
}

func NewCRUD(conn psql.Conn, repo Repo) *CRUD {
	return &CRUD{conn: conn, repo: repo}
}

func (c *CRUD) Create(ctx context.Context, userID uuid.UUID, request Session) (Session, error) {
	request.Key = generateKey()
	request.Owner = userID
	if err := request.Validate(); err != nil {
		return Session{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}

	created, err := c.repo.Create(ctx, c.conn, request)
	if err != nil {
		return Session{}, fmt.Errorf("failed to create session: %w", err)
	}

	return created[0], nil
}

func (c *CRUD) Fetch(ctx context.Context, Key string) (Session, error) {

	fetched, err := c.repo.FetchOne(ctx, c.conn, Lookup{Key: Key})
	if err != nil {
		return Session{}, fmt.Errorf("failed to fetch session: %w", err)
	}
	return fetched, nil
}

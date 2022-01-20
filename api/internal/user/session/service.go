package session

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Repo interface {
	Create(ctx context.Context, requests ...Session) ([]Session, error)
	FetchOne(ctx context.Context, lookup Lookup) (Session, error)
}

type CRUD struct {
	repo Repo
}

func NewCRUD(repo Repo) *CRUD {
	return &CRUD{repo: repo}
}

func (c *CRUD) Create(ctx context.Context, userID uuid.UUID) (Session, error) {
	created, err := c.repo.Create(ctx, Session{
		Key:   NewKey(),
		Owner: userID,
	})
	if err != nil {
		return Session{}, fmt.Errorf("failed to create session: %w", err)
	}

	return created[0], nil
}

func (c *CRUD) Fetch(ctx context.Context, key string) (Session, error) {
	if key == "" {
		return Session{}, ErrNotFound
	}
	return c.repo.FetchOne(ctx, Lookup{Key: key})
}

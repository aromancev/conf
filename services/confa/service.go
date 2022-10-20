package confa

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Repo interface {
	Create(ctx context.Context, requests ...Confa) ([]Confa, error)
	Fetch(ctx context.Context, lookup Lookup) ([]Confa, error)
	UpdateOne(ctx context.Context, lookup Lookup, request Mask) (Confa, error)
}

type CRUD struct {
	repo Repo
}

func NewCRUD(repo Repo) *CRUD {
	return &CRUD{repo: repo}
}

func (c *CRUD) Create(ctx context.Context, userID uuid.UUID, request Confa) (Confa, error) {
	request.ID = uuid.New()
	request.Owner = userID
	if request.Handle == "" {
		request.Handle = strings.Split(request.ID.String(), "-")[4]
	}
	created, err := c.repo.Create(ctx, request)
	if err != nil {
		return Confa{}, fmt.Errorf("failed to create confa: %w", err)
	}
	return created[0], nil
}

func (c *CRUD) Update(ctx context.Context, userID uuid.UUID, lookup Lookup, request Mask) (Confa, error) {
	lookup.Owner = userID
	return c.repo.UpdateOne(ctx, lookup, request)
}

func (c *CRUD) Fetch(ctx context.Context, lookup Lookup) ([]Confa, error) {
	return c.repo.Fetch(ctx, lookup)
}

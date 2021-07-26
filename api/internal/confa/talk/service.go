package talk

import (
	"context"
	"fmt"

	"github.com/aromancev/confa/internal/confa"

	"github.com/google/uuid"
)

type Repo interface {
	Create(ctx context.Context, requests ...Talk) ([]Talk, error)
	Fetch(ctx context.Context, lookup Lookup) ([]Talk, error)
	FetchOne(ctx context.Context, lookup Lookup) (Talk, error)
}

type ConfaRepo interface {
	FetchOne(ctx context.Context, lookup confa.Lookup) (confa.Confa, error)
}

type CRUD struct {
	repo   Repo
	confas ConfaRepo
}

func NewCRUD(repo Repo, confas ConfaRepo) *CRUD {
	return &CRUD{repo: repo, confas: confas}
}

func (c *CRUD) Create(ctx context.Context, userID uuid.UUID, request Talk) (Talk, error) {
	request.ID = uuid.New()
	request.Owner = userID
	request.Speaker = userID
	if request.Handle == "" {
		request.Handle = request.ID.String()
	}
	if err := request.Validate(); err != nil {
		return Talk{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	conf, err := c.confas.FetchOne(ctx, confa.Lookup{ID: request.Confa})
	if err != nil {
		return Talk{}, fmt.Errorf("failed to fetch confa: %w", err)
	}
	if conf.Owner != userID {
		return Talk{}, ErrPermissionDenied
	}
	created, err := c.repo.Create(ctx, request)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to create talk: %w", err)
	}
	return created[0], nil
}

func (c *CRUD) Fetch(ctx context.Context, lookup Lookup) ([]Talk, error) {
	return c.repo.Fetch(ctx, lookup)
}

func (c *CRUD) Start(ctx context.Context, userID, talkID uuid.UUID) error {
	talk, err := c.repo.FetchOne(ctx, Lookup{ID: talkID})
	if err != nil {
		return fmt.Errorf("failed to fetch talk: %w", err)
	}
	if talk.Owner != userID {
		return ErrPermissionDenied
	}
	return nil

}

package clap

import (
	"context"
	"fmt"

	"github.com/aromancev/confa/confa/talk"
	"github.com/google/uuid"
)

type Repo interface {
	CreateOrUpdate(ctx context.Context, request Clap) (uuid.UUID, error)
	Aggregate(ctx context.Context, lookup Lookup) (uint64, error)
}

type TalkRepo interface {
	FetchOne(ctx context.Context, lookup talk.Lookup) (talk.Talk, error)
}

type CRUD struct {
	repo     Repo
	talkRepo TalkRepo
}

func NewCRUD(repo Repo, talkRepo TalkRepo) *CRUD {
	return &CRUD{repo: repo, talkRepo: talkRepo}
}

func (c *CRUD) CreateOrUpdate(ctx context.Context, userID, talkID uuid.UUID, value uint) (uuid.UUID, error) {
	tlk, err := c.talkRepo.FetchOne(ctx, talk.Lookup{ID: talkID})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to fetch talk: %w", err)
	}
	return c.repo.CreateOrUpdate(ctx, Clap{
		ID:      uuid.New(),
		Owner:   userID,
		Confa:   tlk.Confa,
		Talk:    tlk.ID,
		Speaker: tlk.Owner,
		Value:   value,
	})
}

func (c *CRUD) Aggregate(ctx context.Context, lookup Lookup, userID uuid.UUID) (Claps, error) {
	claps, err := c.repo.Aggregate(ctx, lookup)
	if err != nil {
		return Claps{}, err
	}
	lookup.Owner = userID
	userClaps, err := c.repo.Aggregate(ctx, lookup)
	if err != nil {
		return Claps{}, err
	}
	cl := Claps{int(claps), int(userClaps)}
	return cl, nil
}

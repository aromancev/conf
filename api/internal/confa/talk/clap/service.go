package clap

import (
	"context"
	"errors"
	"fmt"
	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/aromancev/confa/internal/platform/psql"
	"github.com/google/uuid"
)

type Repo interface {
	CreateOrUpdate(ctx context.Context, execer psql.Execer, request Clap) (Clap, error)
	Aggregate(ctx context.Context, queryer psql.Queryer, lookup Lookup) (int, error)
}
type TalkRepo interface {
	FetchOne(ctx context.Context, queryer psql.Queryer, lookup talk.Lookup) (talk.Talk, error)
}
type CRUD struct {
	conn     psql.Conn
	repo     Repo
	talkRepo TalkRepo
}

func NewCRUD(conn psql.Conn, repo Repo, talkRepo TalkRepo) *CRUD {
	return &CRUD{conn: conn, repo: repo, talkRepo: talkRepo}
}

func (c *CRUD) CreateOrUpdate(ctx context.Context, ownerID uuid.UUID, request Clap) (Clap, error) {
	request.ID = uuid.New()
	request.Owner = ownerID
	fetchedTalk, err := c.talkRepo.FetchOne(ctx, c.conn, talk.Lookup{ID: request.Talk})
	switch {
		case errors.Is(err, talk.ErrNotFound):
			return Clap{}, ErrValidation
		case err != nil:
			return Clap{}, fmt.Errorf("failed to create clap: %w", err)
	}
	request.Confa = fetchedTalk.Confa
	request.Speaker = fetchedTalk.Speaker
	if err := request.Validate(); err != nil {
		return Clap{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	created, err := c.repo.CreateOrUpdate(ctx, c.conn, request)
	if err != nil {
		return Clap{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	return created, nil
}

func (c *CRUD) Aggregate(ctx context.Context, lookup Lookup) (int, error) {
	aggregated, err := c.repo.Aggregate(ctx, c.conn, lookup)
	if err != nil {
		return 0, fmt.Errorf("failed to aggregate clap: %w", err)
	}
	return aggregated, nil
}

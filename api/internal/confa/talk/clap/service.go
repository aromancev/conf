package clap

import (
	"context"
	"fmt"
	"github.com/google/uuid"

	"github.com/aromancev/confa/internal/platform/psql"
)

type Repo interface {
	CreateOrUpdate(ctx context.Context, execer psql.Execer, request Clap) error
	Aggregate(ctx context.Context, queryer psql.Queryer, lookup Lookup) (int, error)
}


type CRUD struct {
	conn      psql.Conn
	repo      Repo
}

func NewCRUD(conn psql.Conn, repo Repo) *CRUD {
	return &CRUD{conn: conn, repo: repo}
}

func (c *CRUD) CreateOrUpdate(ctx context.Context, ownerID uuid.UUID, request Clap) error {
	request.ID = uuid.New()
	request.Owner = ownerID
	request.Speaker = ownerID
	if err := request.Validate(); err != nil {
		return fmt.Errorf("%w: %s", ErrValidation, err)
	}
	err := c.repo.CreateOrUpdate(ctx, c.conn, request)
	if err != nil {
		return fmt.Errorf("failed to create clap: %w", err)
	}
	return nil
}

func (c *CRUD) Aggregate(ctx context.Context, lookup Lookup) (int, error) {
	aggregated, err := c.repo.Aggregate(ctx, c.conn, lookup)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch clap: %w", err)
	}
	return aggregated, nil
}

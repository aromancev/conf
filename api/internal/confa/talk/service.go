package talk

import (
	"context"
	"fmt"

	"github.com/aromancev/confa/internal/confa"

	"github.com/google/uuid"

	"github.com/aromancev/confa/internal/platform/psql"
)

type Repo interface {
	Create(ctx context.Context, execer psql.Execer, requests ...Talk) ([]Talk, error)
	FetchOne(ctx context.Context, queryer psql.Queryer, lookup Lookup) (Talk, error)
}
type ConfaCRUD interface {
	Fetch(ctx context.Context, ID uuid.UUID) (confa.Confa, error)
}

type CRUD struct {
	conn      psql.Conn
	repo      Repo
	confaCRUD ConfaCRUD
}

func NewCRUD(conn psql.Conn, repo Repo, confaCRUD ConfaCRUD) *CRUD {
	return &CRUD{conn: conn, repo: repo, confaCRUD: confaCRUD}
}

func (c *CRUD) Create(ctx context.Context, confaID, ownerID uuid.UUID, request Talk) (Talk, error) {
	request.ID = uuid.New()
	request.Confa = confaID
	request.Owner = ownerID
	request.Speaker = ownerID
	if request.Handle == "" {
		request.Handle = request.ID.String()
	}
	if err := request.Validate(); err != nil {
		return Talk{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	fetchedConfa, err := c.confaCRUD.Fetch(ctx, confaID)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to fetch confa: %w", err)
	}

	if fetchedConfa.Owner != ownerID {
		return Talk{}, ErrPermissionDenied
	}

	created, err := c.repo.Create(ctx, c.conn, request)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to create talk: %w", err)
	}

	return created[0], nil
}

func (c *CRUD) Fetch(ctx context.Context, id uuid.UUID) (Talk, error) {
	fetched, err := c.repo.FetchOne(ctx, c.conn, Lookup{ID: id})
	if err != nil {
		return Talk{}, fmt.Errorf("failed to fetch talk: %w", err)
	}
	return fetched, nil
}

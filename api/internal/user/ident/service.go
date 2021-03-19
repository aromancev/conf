package ident

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"github.com/aromancev/confa/internal/platform/backoff"
	"github.com/aromancev/confa/internal/platform/psql"
	"github.com/aromancev/confa/internal/user"
)

type Repo interface {
	Create(ctx context.Context, execer psql.Execer, requests ...Ident) ([]Ident, error)
	Fetch(ctx context.Context, queryer psql.Queryer, lookup Lookup) ([]Ident, error)
	FetchOne(ctx context.Context, queryer psql.Queryer, lookup Lookup) (Ident, error)
}

type UserRepo interface {
	Create(ctx context.Context, execer psql.Execer, requests ...user.User) ([]user.User, error)
}

type CRUD struct {
	conn   psql.Conn
	idents Repo
	users  UserRepo
}

func NewCRUD(conn psql.Conn, idents Repo, users UserRepo) *CRUD {
	return &CRUD{conn: conn, idents: idents, users: users}
}

func (s *CRUD) GetOrCreate(ctx context.Context, with Ident) (uuid.UUID, error) {
	usr := user.User{
		ID: uuid.New(),
	}
	with.ID = uuid.New()
	with.Owner = usr.ID
	if err := with.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", err, ErrValidation)
	}

	try := func() (uuid.UUID, error) {
		ident, err := s.idents.FetchOne(ctx, s.conn, Lookup{Matching: []Ident{with}})
		switch {
		case err == nil:
			return ident.Owner, nil

		case errors.Is(err, ErrNotFound):
			err := psql.Tx(ctx, s.conn, func(ctx context.Context, tx pgx.Tx) error {
				_, err := s.users.Create(ctx, tx, usr)
				if err != nil {
					return err
				}
				_, err = s.idents.Create(ctx, tx, with)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
			}
			return usr.ID, nil

		default:
			return uuid.Nil, fmt.Errorf("failed to fetch idents: %w", err)
		}
	}

	bo := backoff.Backoff{
		Factor: 1.2,
		Jitter: true,
		Min:    time.Millisecond,
		Max:    50 * time.Millisecond,
	}
	var err error
	var userID uuid.UUID
	for i := 0; i < 3; i++ {
		userID, err = try()
		switch {
		case err == nil:
			return userID, err
		case errors.Is(err, ErrDuplicatedEntry):
			time.Sleep(bo.ForAttempt(float64(i)))
			continue
		case err != nil:
			return uuid.Nil, err
		}
	}
	return uuid.Nil, fmt.Errorf("retries exceeded: %w", err)
}

package iam

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"github.com/aromancev/confa/internal/platform/psql"
)

type IdentRepo interface {
	Create(ctx context.Context, execer psql.Execer, requests ...Ident) ([]Ident, error)
	Fetch(ctx context.Context, queryer psql.Queryer, lookup IdentLookup) ([]Ident, error)
	FetchOne(ctx context.Context, queryer psql.Queryer, lookup IdentLookup) (Ident, error)
}

type UserRepo interface {
	Create(ctx context.Context, execer psql.Execer, requests ...User) ([]User, error)
}

type CRUD struct {
	conn   pgx.Tx
	idents IdentRepo
	users  UserRepo
}

func NewCRUD(conn pgx.Tx, idents IdentRepo, users UserRepo) *CRUD {
	return &CRUD{conn: conn, idents: idents, users: users}
}

func (s *CRUD) GetOrCreate(ctx context.Context, with Ident) (uuid.UUID, error) {
	user := User{
		ID: uuid.New(),
	}
	with.ID = uuid.New()
	with.Owner = user.ID
	if err := with.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", err, ErrValidation)
	}

	try := func() (uuid.UUID, error) {
		ident, err := s.idents.FetchOne(ctx, s.conn, IdentLookup{Matching: []Ident{with}})
		switch {
		case err == nil:
			return ident.Owner, nil

		case errors.Is(err, ErrNotFound):
			err := psql.Tx(ctx, s.conn, func(ctx context.Context, tx pgx.Tx) error {
				_, err := s.users.Create(ctx, tx, user)
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
			return user.ID, nil

		default:
			return uuid.Nil, fmt.Errorf("failed to fetch idents: %w", err)
		}
	}

	var err error
	var userID uuid.UUID
	for i := 0; i < 3; i++ {
		userID, err = try()
		switch {
		case err == nil:
			return userID, err
		case errors.Is(err, ErrDuplicatedEntry):
			continue
		case err != nil:
			return uuid.Nil, err
		}
	}
	return uuid.Nil, fmt.Errorf("retries exceeded: %w", err)
}

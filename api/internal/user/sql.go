package user

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/aromancev/confa/internal/platform/psql"
)

const (
	batchLimit = 500
)

type SQL struct {
}

func NewSQL() *SQL {
	return &SQL{}
}

func (s *SQL) Create(ctx context.Context, execer psql.Execer, requests ...User) ([]User, error) {
	if len(requests) == 0 {
		return nil, errors.New("trying to create zero objects")
	}
	if len(requests) > batchLimit {
		return nil, fmt.Errorf("trying to create more than %d", batchLimit)
	}

	for i, r := range requests {
		if err := r.Validate(); err != nil {
			return nil, fmt.Errorf("invalid request [%d]: %w", i, err)
		}
	}

	now := psql.Now()
	for i := range requests {
		requests[i].CreatedAt = now
	}

	q := sq.Insert("users").Columns("id", "created_at")
	for _, r := range requests {
		q = q.Values(r.ID, r.CreatedAt)
	}
	q = q.PlaceholderFormat(sq.Dollar)
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	_, err = execer.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (s *SQL) Fetch(ctx context.Context, queryer psql.Queryer, lookup Lookup) ([]User, error) {
	q := sq.Select("id", "created_at")
	q = q.From("users")
	if lookup.ID != uuid.Nil {
		q = q.Where(sq.Eq{"id": lookup.ID})
	}

	q = q.Limit(batchLimit)
	q = q.PlaceholderFormat(sq.Dollar)
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := queryer.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	var users []User
	for rows.Next() {
		var c User
		err := rows.Scan(
			&c.ID,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		c.CreatedAt = c.CreatedAt.UTC()
		users = append(users, c)
	}

	return users, nil
}

func (s *SQL) FetchOne(ctx context.Context, queryer psql.Queryer, lookup Lookup) (User, error) {
	users, err := s.Fetch(ctx, queryer, lookup)
	if err != nil {
		return User{}, err
	}
	if len(users) == 0 {
		return User{}, ErrNotFound
	}
	if len(users) > 1 {
		return User{}, ErrUnexpectedResult
	}
	return users[0], nil
}

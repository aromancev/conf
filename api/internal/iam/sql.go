package iam

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"github.com/aromancev/confa/internal/platform/psql"
)

const (
	batchLimit = 500
)

type IdentSQL struct {
}

func NewIdentSQL() *IdentSQL {
	return &IdentSQL{}
}

func (s *IdentSQL) Create(ctx context.Context, execer psql.Execer, requests ...Ident) ([]Ident, error) {
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

	now := time.Now().Round(time.Millisecond).UTC()
	for i := range requests {
		requests[i].CreatedAt = now
	}

	q := sq.Insert("idents").Columns("id", "owner", "platform", "value", "created_at")
	for _, r := range requests {
		q = q.Values(r.ID, r.Owner, r.Platform, r.Value, r.CreatedAt)
	}
	q = q.PlaceholderFormat(sq.Dollar)
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = execer.Exec(ctx, query, args...)
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr):
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return nil, ErrDuplicatedEntry
		}
		return nil, err

	case err != nil:
		return nil, err
	}
	return requests, nil
}

func (s *IdentSQL) Fetch(ctx context.Context, queryer psql.Queryer, lookup IdentLookup) ([]Ident, error) {
	q := sq.Select("id", "owner", "platform", "value", "created_at")
	q = q.From("idents")
	if lookup.ID != uuid.Nil {
		q = q.Where(sq.Eq{"id": lookup.ID})
	}
	if lookup.Owner != uuid.Nil {
		q = q.Where(sq.Eq{"owner": lookup.Owner})
	}
	if len(lookup.Matching) != 0 {
		matches := make(sq.Or, len(lookup.Matching))
		for i, m := range lookup.Matching {
			matches[i] = sq.Eq{"platform": m.Platform, "value": m.Value}
		}
		q = q.Where(matches)
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
	var idents []Ident
	for rows.Next() {
		var c Ident
		err := rows.Scan(
			&c.ID,
			&c.Owner,
			&c.Platform,
			&c.Value,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		c.CreatedAt = c.CreatedAt.UTC()
		idents = append(idents, c)
	}

	return idents, nil
}

func (s *IdentSQL) FetchOne(ctx context.Context, queryer psql.Queryer, lookup IdentLookup) (Ident, error) {
	idents, err := s.Fetch(ctx, queryer, lookup)
	if err != nil {
		return Ident{}, err
	}
	if len(idents) == 0 {
		return Ident{}, ErrNotFound
	}
	if len(idents) > 1 {
		return Ident{}, ErrUnexpectedResult
	}
	return idents[0], nil
}

type UserSQL struct {
}

func NewUserSQL() *UserSQL {
	return &UserSQL{}
}

func (s *UserSQL) Create(ctx context.Context, execer psql.Execer, requests ...User) ([]User, error) {
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

	now := time.Now().Round(time.Millisecond).UTC()
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

func (s *UserSQL) Fetch() {

}

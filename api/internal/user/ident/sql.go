package ident

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

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

func (s *SQL) Create(ctx context.Context, execer psql.Execer, requests ...Ident) ([]Ident, error) {
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
		if pgErr.Code == pgerrcode.UniqueViolation {
			return nil, ErrDuplicatedEntry
		}
		return nil, err

	case err != nil:
		return nil, err
	}
	return requests, nil
}

func (s *SQL) Fetch(ctx context.Context, queryer psql.Queryer, lookup Lookup) ([]Ident, error) {
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

func (s *SQL) FetchOne(ctx context.Context, queryer psql.Queryer, lookup Lookup) (Ident, error) {
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

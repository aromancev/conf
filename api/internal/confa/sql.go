package confa

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

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

func (s *SQL) Create(ctx context.Context, execer psql.Execer, requests ...Confa) ([]Confa, error) {
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

	q := sq.Insert("confas").Columns("id", "owner", "handle", "created_at")
	for _, r := range requests {
		q = q.Values(r.ID, r.Owner, r.Handle, r.CreatedAt)
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

func (s *SQL) Fetch(ctx context.Context, queryer psql.Queryer, lookup Lookup) ([]Confa, error) {
	q := sq.Select("id", "owner", "handle", "created_at").From("confas")
	if lookup.ID != uuid.Nil {
		q = q.Where(sq.Eq{"id": lookup.ID})
	}
	if lookup.Owner != uuid.Nil {
		q = q.Where(sq.Eq{"owner": lookup.Owner})
	}
	if lookup.Handle != "" {
		q = q.Where(sq.Eq{"handle": lookup.Handle})
	}
	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}
	q = q.OrderBy("id")
	q = q.Where(sq.Gt{"id": lookup.From})
	q = q.Limit(lookup.Limit)
	q = q.PlaceholderFormat(sq.Dollar)
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := queryer.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	var confas []Confa
	for rows.Next() {
		var c Confa
		err := rows.Scan(
			&c.ID,
			&c.Owner,
			&c.Handle,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		c.CreatedAt = c.CreatedAt.UTC()
		confas = append(confas, c)
	}

	return confas, nil
}

func (s *SQL) FetchOne(ctx context.Context, queryer psql.Queryer, lookup Lookup) (Confa, error) {
	confas, err := s.Fetch(ctx, queryer, lookup)
	if err != nil {
		return Confa{}, err
	}
	if len(confas) == 0 {
		return Confa{}, ErrNotFound
	}
	if len(confas) > 1 {
		return Confa{}, ErrUnexpectedResult
	}
	return confas[0], nil
}

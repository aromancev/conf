package session

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

func (s *SQL) Create(ctx context.Context, execer psql.Execer, requests ...Session) ([]Session, error) {
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

	q := sq.Insert("sessions").Columns("key", "owner", "created_at")
	for _, r := range requests {
		q = q.Values(r.Key, r.Owner, r.CreatedAt)
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

func (s *SQL) Fetch(ctx context.Context, queryer psql.Queryer, lookup Lookup) ([]Session, error) {
	q := sq.Select("key", "owner", "created_at").From("sessions")
	if lookup.Key != "" {
		q = q.Where(sq.Eq{"Key": lookup.Key})
	}
	if lookup.Owner != uuid.Nil {
		q = q.Where(sq.Eq{"owner": lookup.Owner})
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
	var sessions []Session
	for rows.Next() {
		var c Session
		err := rows.Scan(
			&c.Key,
			&c.Owner,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		c.CreatedAt = c.CreatedAt.UTC()
		sessions = append(sessions, c)
	}

	return sessions, nil
}

func (s *SQL) FetchOne(ctx context.Context, queryer psql.Queryer, lookup Lookup) (Session, error) {
	sessions, err := s.Fetch(ctx, queryer, lookup)
	if err != nil {
		return Session{}, err
	}
	if len(sessions) == 0 {
		return Session{}, ErrNotFound
	}
	if len(sessions) > 1 {
		return Session{}, ErrUnexpectedResult
	}
	return sessions[0], nil
}

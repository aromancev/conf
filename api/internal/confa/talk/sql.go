package talk

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

func (s *SQL) Create(ctx context.Context, execer psql.Execer, requests ...Talk) ([]Talk, error) {
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

	q := sq.Insert("talks").Columns("id", "confa", "handle", "created_at")
	for _, r := range requests {
		q = q.Values(r.ID, r.Confa, r.Handle, r.CreatedAt)
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

func (s *SQL) Fetch(ctx context.Context, queryer psql.Queryer, lookup Lookup) ([]Talk, error) {
	q := sq.Select("id", "confa", "handle", "created_at").From("talks")
	if lookup.ID != uuid.Nil {
		q = q.Where(sq.Eq{"id": lookup.ID})
	}
	if lookup.Confa != uuid.Nil {
		q = q.Where(sq.Eq{"confa": lookup.Confa})
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
	var talks []Talk
	for rows.Next() {
		var t Talk
		err := rows.Scan(
			&t.ID,
			&t.Confa,
			&t.Handle,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		t.CreatedAt = t.CreatedAt.UTC()
		talks = append(talks, t)
	}

	return talks, nil
}

func (s *SQL) FetchOne(ctx context.Context, queryer psql.Queryer, lookup Lookup) (Talk, error) {
	talks, err := s.Fetch(ctx, queryer, lookup)
	if err != nil {
		return Talk{}, err
	}
	if len(talks) == 0 {
		return Talk{}, ErrNotFound
	}
	if len(talks) > 1 {
		return Talk{}, ErrUnexpectedResult
	}
	return talks[0], nil
}

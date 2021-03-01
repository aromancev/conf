package confa

import (
	"context"
	"errors"
	"fmt"
	"time"

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

	now := time.Now().UTC()
	for i := range requests {
		requests[i].CreatedAt = now
	}

	b := psql.NewValuesBuilder()
	for _, r := range requests {
		b.WriteRow(r.ID, r.Owner, r.Name, r.CreatedAt)
	}
	query, args := b.Query()
	_, err := execer.Exec(
		ctx,
		`
			INSERT INTO confas
			(id, owner, name, created_at)
			VALUES
		`+query,
		args...,
	)
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (s *SQL) Fetch(ctx context.Context, queryer psql.Queryer, lookup Lookup) ([]Confa, error) {
	b := psql.NewConditionBuilder("AND")
	if lookup.ID != uuid.Nil {
		b.Eq("id", lookup.ID)
	}
	if lookup.Owner != uuid.Nil {
		b.Eq("owner", lookup.Owner)
	}
	query, args := b.Query()
	rows, err := queryer.Query(
		ctx,
		`
		SELECT id, owner, name, created_at
		FROM confas
		WHERE
		`+" "+query,
		args...,
	)
	if err != nil {
		return nil, err
	}

	var confas []Confa
	for rows.Next() {
		var c Confa
		err := rows.Scan(
			&c.ID,
			&c.Owner,
			&c.Name,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
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
		return Confa{}, ErrNoRows
	}
	if len(confas) > 1 {
		return Confa{}, ErrMultipleRows
	}
	return confas[0], nil
}

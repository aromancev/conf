package confa

import (
	"context"
	"errors"
	"fmt"
	"time"

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
		b.WriteRow(r.ID, r.Owner, r.Tag, r.CreatedAt)
	}
	query, args := b.Query()
	_, err := execer.ExecContext(
		ctx,
		`
			INSERT INTO confas
			(id, owner, tag, created_at)
			VALUES
		`+query,
		args...,
	)
	if err != nil {
		return nil, err
	}
	return requests, nil
}

package clap

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/aromancev/confa/internal/platform/psql"
	"github.com/google/uuid"
)

type SQL struct {
}

func NewSQL() *SQL {
	return &SQL{}
}

func (s *SQL) CreateOrUpdate(ctx context.Context, execer psql.Execer, request Clap) error {
	err := request.Validate()
	if err != nil {
		return err
	}
	q := sq.Insert("claps").Columns("owner", "speaker", "confa", "talk", "claps")
	q = q.Values(request.Owner, request.Speaker, request.Confa, request.Talk, request.Claps)
	q = q.Suffix("ON CONFLICT ON CONSTRAINT unique_owner_talk DO UPDATE SET claps = ?", request.Claps)
	q = q.PlaceholderFormat(sq.Dollar)
	query, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = execer.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQL) Aggregate(ctx context.Context, queryer psql.Queryer, lookup Lookup) (int, error) {
	q := sq.Select("SUM(claps)").From("claps")
	if lookup.Speaker != uuid.Nil {
		q = q.Where(sq.Eq{"speaker": lookup.Speaker})
	}
	if lookup.Confa != uuid.Nil {
		q = q.Where(sq.Eq{"confa": lookup.Confa})
	}
	if lookup.Talk != uuid.Nil {
		q = q.Where(sq.Eq{"talk": lookup.Talk})
	}
	q = q.PlaceholderFormat(sq.Dollar)
	query, args, err := q.ToSql()
	if err != nil {
		return 0, err
	}

	rows, err := queryer.Query(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	var claps int
	for rows.Next() {
		err := rows.Scan(
			&claps,
		)
		if err != nil {
			return 0, err
		}
	}

	return claps, nil
}

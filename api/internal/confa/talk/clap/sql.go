package clap

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	sq "github.com/Masterminds/squirrel"
	"github.com/aromancev/confa/internal/platform/psql"
)

type SQL struct {
}

func NewSQL() *SQL {
	return &SQL{}
}

func (s *SQL) CreateOrUpdate(ctx context.Context, execer psql.Execer, request Clap) error {
	err := request.Validate()
	if err != nil {
		return fmt.Errorf("invalid request : %w", err)
	}

	q := sq.Insert("claps").Columns("id", "owner", "speaker", "confa", "talk", "claps")
	q = q.Values(request.ID, request.Owner, request.Speaker, request.Confa, request.Talk, request.Claps)
	q = q.Suffix("ON CONFLICT ON CONSTRAINT unique_owner_talk DO UPDATE SET claps = ?", request.Claps)
	q = q.PlaceholderFormat(sq.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = execer.Exec(ctx, query, args...)
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr):
		if pgErr.Code == pgerrcode.UniqueViolation {
			return ErrDuplicatedEntry
		}
		return err

	case err != nil:
		return err
	}
	return nil
}

func (s *SQL) Aggregate(ctx context.Context, queryer psql.Queryer, lookup Lookup) (int, error) {
	q := sq.Select("claps").From("claps")
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
		var c Clap
		err := rows.Scan(
			&c.Claps,
		)
		if err != nil {
			return 0, err
		}
		claps += int(c.Claps)
	}

	return claps, nil
}

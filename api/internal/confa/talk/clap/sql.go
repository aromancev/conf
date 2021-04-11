package clap

import (
	"context"
	"errors"
	"fmt"

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

func (s *SQL) CreateOrUpdate(ctx context.Context, execer psql.Execer, request Clap) (Clap, error) {
	err := request.Validate()
	if err != nil {
		return Clap{}, fmt.Errorf("invalid request : %w", err)
	}

	q := sq.Insert("claps").Columns("id", "owner", "speaker", "confa", "talk", "claps")
	q = q.Values(request.ID, request.Owner, request.Speaker, request.Confa, request.Talk, request.Claps)
	q = q.Suffix("ON CONFLICT ON CONSTRAINT unique_owner_talk DO UPDATE SET claps = ?", request.Claps)
	q = q.PlaceholderFormat(sq.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return Clap{}, err
	}
	_, err = execer.Exec(ctx, query, args...)
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr):
		if pgErr.Code == pgerrcode.UniqueViolation {
			return Clap{}, ErrDuplicatedEntry
		}
		return Clap{}, err

	case err != nil:
		return Clap{}, err
	}
	return request, nil
}

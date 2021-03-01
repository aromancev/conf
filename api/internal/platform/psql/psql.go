package psql

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Queryer interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type Execer interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type Config struct {
	Host, Port, User, Password, Database string
}

func New(ctx context.Context, c Config) (*pgxpool.Pool, error) {
	conn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s",
		c.Host, c.Port, c.User, c.Password, c.Database,
	)
	return pgxpool.Connect(ctx, conn)
}

// ValuesBuilder can be used to build queries for multiple insert with placeholders.
type ValuesBuilder struct {
	sql  strings.Builder
	args []interface{}
}

func NewValuesBuilder() *ValuesBuilder {
	return &ValuesBuilder{}
}

func (b *ValuesBuilder) WriteRow(args ...interface{}) {
	if len(b.args) != 0 {
		b.sql.WriteString(",")
	}
	b.sql.WriteString("(")
	for i := range args {
		if i != 0 {
			b.sql.WriteString(",")
		}
		b.sql.WriteString(fmt.Sprintf("$%d", len(b.args)+i+1))
	}
	b.sql.WriteString(")")
	b.args = append(b.args, args...)
}

func (b *ValuesBuilder) Query() (string, []interface{}) {
	return b.sql.String(), b.args
}

// ConditionBuilder can be used to build queries with long dynamic conditions.
type ConditionBuilder struct {
	sql  strings.Builder
	args []interface{}
	sep  string
}

func NewConditionBuilder(separator string) *ConditionBuilder {
	return &ConditionBuilder{sep: separator}
}

func (b *ConditionBuilder) Eq(column string, val interface{}) {
	if len(b.args) != 0 {
		b.sql.WriteString(" " + b.sep + " ")
	}
	b.sql.WriteString(fmt.Sprintf("%s = $%d", column, len(b.args)+1))
	b.args = append(b.args, val)
}

func (b *ConditionBuilder) Query() (string, []interface{}) {
	return b.sql.String(), b.args
}

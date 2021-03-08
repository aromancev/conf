package psql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Queryer interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type Execer interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type Config struct {
	Host, User, Password, Database string
	Port                           uint16
}

func New(ctx context.Context, c Config) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, err
	}

	cfg.ConnConfig = &pgx.ConnConfig{
		Config: pgconn.Config{
			Host:           c.Host,
			Port:           c.Port,
			Database:       c.Database,
			User:           c.User,
			Password:       c.Password,
			ConnectTimeout: 10 * time.Second,
		},
		Logger:   &logger{},
		LogLevel: pgx.LogLevelWarn,
	}
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 10 * time.Minute
	cfg.MaxConns = 10
	cfg.LazyConnect = true

	return pgxpool.ConnectConfig(ctx, cfg)
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

type logger struct{}

func (l logger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	var event *zerolog.Event
	switch level {
	case pgx.LogLevelError:
		event = log.Ctx(ctx).Error()
	case pgx.LogLevelWarn:
		event = log.Ctx(ctx).Warn()
	default:
		event = log.Ctx(ctx).Info()
	}

	event.Fields(data).Msg(msg)
}

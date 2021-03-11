package psql

import (
	"context"
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

type Txer interface {
	Begin(ctx context.Context) (pgx.Tx, error)
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

func Tx(ctx context.Context, conn Txer, f func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	err = f(ctx, tx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func Now() time.Time {
	return time.Now().UTC().Round(time.Microsecond)
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

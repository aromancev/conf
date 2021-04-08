package psql

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
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

type Conn interface {
	Queryer
	Execer
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

	cfg.ConnConfig.Config.Host = c.Host
	cfg.ConnConfig.Config.Port = c.Port
	cfg.ConnConfig.Config.Database = c.Database
	cfg.ConnConfig.Config.User = c.User
	cfg.ConnConfig.Config.Password = c.Password
	cfg.ConnConfig.Config.ConnectTimeout = 10 * time.Second
	cfg.ConnConfig.Logger = &logger{}
	cfg.ConnConfig.LogLevel = pgx.LogLevelWarn
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 10 * time.Minute
	cfg.MaxConns = 10
	cfg.LazyConnect = true

	return pgxpool.ConnectConfig(ctx, cfg)
}

func Tx(ctx context.Context, conn Conn, f func(ctx context.Context, tx pgx.Tx) error) error {
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

type FS interface {
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}

type MigratorFS struct {
	fs FS
}

func NewMigratorFS(f FS) *MigratorFS {
	return &MigratorFS{fs: f}
}

func (m MigratorFS) ReadDir(dirname string) ([]os.FileInfo, error) {
	entries, err := m.fs.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	infos := make([]os.FileInfo, len(entries))
	for i, e := range entries {
		info, err := e.Info()
		if err != nil {
			return nil, err
		}
		infos[i] = info
	}
	return infos, nil
}

func (m MigratorFS) ReadFile(filename string) ([]byte, error) {
	return m.fs.ReadFile(filename)
}

func (m MigratorFS) Glob(pattern string) ([]string, error) {
	entries, err := m.fs.ReadDir(".")
	if err != nil {
		return nil, err
	}

	matches := make([]string, 0, len(entries))
	for _, e := range entries {
		ok, err := filepath.Match(pattern, e.Name())
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		matches = append(matches, e.Name())
	}

	return matches, nil
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

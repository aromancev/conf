package migrations

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/tern/migrate"

	"github.com/aromancev/confa/internal/platform/psql"
)

const (
	version = "public.schema_version"
)

//go:embed *.sql
var migrations embed.FS

func NewMigrator(ctx context.Context, conn *pgx.Conn) (*migrate.Migrator, error) {
	return migrate.NewMigratorEx(ctx, conn, version, &migrate.MigratorOptions{
		MigratorFS: psql.NewMigratorFS(migrations),
	})
}

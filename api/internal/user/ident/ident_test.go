package ident

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v4"

	"github.com/aromancev/confa/internal/platform/psql/double"
	"github.com/aromancev/confa/internal/user/migrations"
)

func TestMain(m *testing.M) {
	code := m.Run()
	double.Purge()
	os.Exit(code)
}

func migrate(conn *pgx.Conn) {
	migrator, err := migrations.NewMigrator(context.Background(), conn)
	if err != nil {
		panic(err)
	}
	if err := migrator.LoadMigrations("."); err != nil {
		panic(err)
	}
	if err := migrator.Migrate(context.Background()); err != nil {
		panic(err)
	}
}

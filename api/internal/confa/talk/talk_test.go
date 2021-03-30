package talk

import (
	"context"
	"github.com/aromancev/confa/internal/confa/migrations"
	"github.com/aromancev/confa/internal/platform/psql/double"
	"github.com/jackc/pgx/v4"
	"os"
	"testing"
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

package session

import (
	"os"
	"testing"

	"github.com/aromancev/iam/internal/platform/mongo/double"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMain(m *testing.M) {
	code := m.Run()
	double.Purge()
	os.Exit(code)
}

func dockerMongo(t *testing.T) *mongo.Database {
	t.Helper()

	db := double.NewDocker()
	driver, err := mongodb.WithInstance(db.Client(), &mongodb.Config{
		DatabaseName: db.Name(),
	})
	require.NoError(t, err)
	migrator, err := migrate.NewWithDatabaseInstance("file://../../migrations/iam", db.Name(), driver)
	require.NoError(t, err)
	require.NoError(t, migrator.Up())
	return db
}

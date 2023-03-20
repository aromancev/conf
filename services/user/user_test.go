package user

import (
	"os"
	"testing"

	"github.com/aromancev/confa/internal/platform/mongo/double"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
	migrator, err := migrate.NewWithDatabaseInstance("file://../migrations/iam", db.Name(), driver)
	require.NoError(t, err)
	require.NoError(t, migrator.Up())
	return db
}

func TestPassword(t *testing.T) {
	t.Parallel()

	pwd := Password(uuid.NewString())
	hash, err := pwd.Hash()
	require.NoError(t, err)
	ok, err := pwd.Check(hash)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = Password(uuid.NewString()).Check(hash)
	require.NoError(t, err)
	assert.False(t, ok)
}

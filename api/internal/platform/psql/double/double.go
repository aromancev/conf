package double

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

// NewDocker creates a Postgres container, runs migrations in it and returns a db connection to use.
// IMPORTANT: Always call returned function at the end of tests (if no error returned) to release docker resources.
// To make this work on CI, set DOCKER_HOST and DOCKER_IP env variables.
func NewDocker(t *testing.T, migrationsDir string) (*sql.DB, func()) {
	t.Helper()

	pool, err := dockertest.NewPool(os.Getenv("DOCKER_HOST"))
	require.NoError(t, err)

	resource, err := pool.Run(
		"postgres",
		"12.6-alpine",
		[]string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
		},
	)
	require.NoError(t, err)

	var db *sql.DB
	if err := pool.Retry(func() error {
		conn := fmt.Sprintf(
			"host=localhost port=%s user=confa password=confa dbname=confa sslmode=disable",
			resource.GetPort("5432/tcp"),
		)
		var err error
		db, err = sql.Open("postgres", conn)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		require.NoError(t, err)
	}

	done := func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}

	return db, done
}

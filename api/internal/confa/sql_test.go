package confa

import (
	"context"
	"github.com/aromancev/confa/internal/platform/psql/double"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	migrations = "../migrations"
)

func TestSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker(migrations)
		defer done()
		sql := NewSQL()

		request := Confa{
			ID:     uuid.New(),
			Owner:  uuid.New(),
			Handle: "test",
		}
		created, err := sql.Create(ctx, pg, request)
		require.NoError(t, err)

		fetched, err := sql.Fetch(ctx, pg, Lookup{
			ID:    request.ID,
			Owner: request.Owner,
		})
		require.NoError(t, err)
		assert.Equal(t, created, fetched)
	})
}

//func TestManualSQL(t *testing.T) {
//	t.Parallel()
//
//	ctx := context.Background()
//
//	t.Run("Create", func(t *testing.T) {
//		t.Parallel()
//		conn := "host=localhost port=5432 user=postgres password=postgres dbname=confa sslmode=disable"
//		pg, err := pgxpool.Connect(context.Background(), conn)
//		if err != nil {
//			panic(err)
//		}
//		sql := NewSQL()
//
//		request := Confa{
//			ID:     uuid.New(),
//			Owner:  uuid.New(),
//			Handle: "test",
//		}
//		created, err := sql.Create(ctx, pg, request)
//
//		require.NoError(t, err)
//
//		fetched, err := sql.Fetch(ctx, pg, Lookup{
//			ID:    request.ID,
//			Owner: request.Owner,
//		})
//		require.NoError(t, err)
//		assert.Equal(t, created, fetched)
//	})
//}

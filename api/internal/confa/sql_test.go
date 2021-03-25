package confa

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()
		sql := NewSQL()

		t.Run("String handle", func(t *testing.T) {
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
		t.Run("UUID handle", func(t *testing.T) {
			request := Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: uuid.New().String(),
			}
			_, err := sql.Create(ctx, pg, request)
			require.NoError(t, err)
		})

		t.Run("Duplicated Entry", func(t *testing.T) {
			t.Parallel()

			pg, done := double.NewDocker("", migrate)
			defer done()

			confas := NewSQL()

			conf := Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test",
			}
			_, err := confas.Create(ctx, pg, conf)
			require.NoError(t, err)

			_, err = confas.Create(ctx, pg, conf)
			require.ErrorIs(t, err, ErrDuplicatedEntry)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()

		confas := NewSQL()

		conf := Confa{
			ID:     uuid.New(),
			Owner:  uuid.New(),
			Handle: "test",
		}
		created, err := confas.Create(ctx, pg, conf)
		require.NoError(t, err)

		t.Run("ID", func(t *testing.T) {
			fetched, err := confas.Fetch(ctx, pg, Lookup{
				ID: conf.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("Owner", func(t *testing.T) {
			fetched, err := confas.Fetch(ctx, pg, Lookup{
				Owner: conf.Owner,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
	})

}

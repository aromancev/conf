package record

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMongo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("FetchOrStart", func(t *testing.T) {
		t.Run("Happy", func(t *testing.T) {
			t.Parallel()

			records := NewMongo(dockerMongo(t))

			record := Record{
				ID:   uuid.New(),
				Room: uuid.New(),
			}
			created, err := records.FetchOrStart(ctx, record)
			require.NoError(t, err)
			assert.NotZero(t, created.CreatedAt)
			assert.NotZero(t, created.StartedAt)
			assert.True(t, created.Active)

			fetched, err := records.FetchOne(ctx, Lookup{
				ID: record.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("Returns existing record", func(t *testing.T) {
			t.Parallel()

			records := NewMongo(dockerMongo(t))

			roomID := uuid.New()
			first, err := records.FetchOrStart(ctx, Record{
				ID:   uuid.New(),
				Room: roomID,
			})
			require.NoError(t, err)

			second, err := records.FetchOrStart(ctx, Record{
				ID:   uuid.New(),
				Room: roomID,
			})
			require.NoError(t, err)
			assert.Equal(t, first, second)
		})

		t.Run("Creates new record for different room", func(t *testing.T) {
			t.Parallel()

			records := NewMongo(dockerMongo(t))

			first, err := records.FetchOrStart(ctx, Record{
				ID:   uuid.New(),
				Room: uuid.New(),
			})
			require.NoError(t, err)

			second, err := records.FetchOrStart(ctx, Record{
				ID:   uuid.New(),
				Room: uuid.New(),
			})
			require.NoError(t, err)
			assert.NotEqual(t, first, second)
		})

		t.Run("Creates a new record if existing if not active", func(t *testing.T) {
			t.Parallel()

			records := NewMongo(dockerMongo(t))

			roomID := uuid.New()
			first, err := records.FetchOrStart(ctx, Record{
				ID:   uuid.New(),
				Room: roomID,
			})
			require.NoError(t, err)

			_, err = records.Stop(ctx, Lookup{ID: first.ID})
			require.NoError(t, err)

			second, err := records.FetchOrStart(ctx, Record{
				ID:   uuid.New(),
				Room: roomID,
			})
			require.NoError(t, err)
			assert.NotEqual(t, first, second)
		})

		t.Run("Returns existing record with same key", func(t *testing.T) {
			t.Parallel()

			records := NewMongo(dockerMongo(t))

			roomID := uuid.New()
			first, err := records.FetchOrStart(ctx, Record{
				ID:   uuid.New(),
				Room: roomID,
				Key:  "test",
			})
			require.NoError(t, err)

			second, err := records.FetchOrStart(ctx, Record{
				ID:   uuid.New(),
				Room: roomID,
				Key:  "test",
			})
			require.NoError(t, err)
			assert.Equal(t, first, second)
		})

		t.Run("Returns error if a record with another key is active", func(t *testing.T) {
			t.Parallel()

			records := NewMongo(dockerMongo(t))

			roomID := uuid.New()
			_, err := records.FetchOrStart(ctx, Record{
				ID:   uuid.New(),
				Room: roomID,
				Key:  "rick",
			})
			require.NoError(t, err)

			_, err = records.FetchOrStart(ctx, Record{
				ID:   uuid.New(),
				Room: roomID,
				Key:  "morty",
			})
			assert.ErrorIs(t, err, ErrDuplicateEntry)
		})
	})

	t.Run("Stop", func(t *testing.T) {
		t.Run("Happy", func(t *testing.T) {
			t.Parallel()

			records := NewMongo(dockerMongo(t))

			created, err := records.FetchOrStart(ctx, Record{
				ID:   uuid.New(),
				Room: uuid.New(),
			})
			require.NoError(t, err)

			res, err := records.Stop(ctx, Lookup{ID: created.ID})
			require.NoError(t, err)
			assert.EqualValues(t, 1, res.ModifiedCount)

			fetched, err := records.FetchOne(ctx, Lookup{ID: created.ID})
			require.NoError(t, err)
			assert.NotZero(t, fetched.StoppedAt)
			assert.False(t, fetched.Active)
		})

		t.Run("Stopping stopped returns 0 modified", func(t *testing.T) {
			t.Parallel()

			records := NewMongo(dockerMongo(t))

			created, err := records.FetchOrStart(ctx, Record{
				ID:   uuid.New(),
				Room: uuid.New(),
			})
			require.NoError(t, err)

			res, err := records.Stop(ctx, Lookup{ID: created.ID})
			require.NoError(t, err)
			assert.EqualValues(t, 1, res.ModifiedCount)
			res, err = records.Stop(ctx, Lookup{ID: created.ID})
			require.NoError(t, err)
			assert.EqualValues(t, 0, res.ModifiedCount)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		records := NewMongo(dockerMongo(t))

		record := Record{
			ID:   uuid.New(),
			Room: uuid.New(),
			Key:  uuid.NewString(),
		}
		created, err := records.FetchOrStart(ctx, record)
		require.NoError(t, err)

		t.Run("By ID", func(t *testing.T) {
			fetched, err := records.FetchOne(ctx, Lookup{
				ID: record.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
		t.Run("By key", func(t *testing.T) {
			fetched, err := records.FetchOne(ctx, Lookup{
				Key: record.Key,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
	})
}

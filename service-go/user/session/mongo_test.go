package session

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			sessions := NewMongo(dockerMongo(t))

			created, err := sessions.Create(ctx, Session{
				Key:   NewKey(),
				Owner: uuid.New(),
			})
			require.NoError(t, err)

			fetched, err := sessions.Fetch(ctx, Lookup{
				Key: created[0].Key,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
		t.Run("Duplicates not allowed", func(t *testing.T) {
			t.Parallel()

			sessions := NewMongo(dockerMongo(t))

			_, err := sessions.Create(ctx, Session{
				Key:   "1",
				Owner: uuid.New(),
			})
			require.NoError(t, err)
			_, err = sessions.Create(ctx, Session{
				Key:   "1",
				Owner: uuid.New(),
			})
			require.ErrorIs(t, err, ErrDuplicatedEntry)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			sessions := NewMongo(dockerMongo(t))

			created, err := sessions.Create(ctx, Session{
				Key:   NewKey(),
				Owner: uuid.New(),
			})
			require.NoError(t, err)

			err = sessions.Delete(ctx, Lookup{
				Owner: created[0].Owner,
			})
			require.NoError(t, err)

			fetched, err := sessions.Fetch(ctx, Lookup{
				Key: created[0].Key,
			})
			require.NoError(t, err)
			assert.Equal(t, 0, len(fetched))
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		sessions := NewMongo(dockerMongo(t))
		created, err := sessions.Create(ctx, Session{
			Key:   NewKey(),
			Owner: uuid.New(),
		})
		require.NoError(t, err)

		t.Run("Key", func(t *testing.T) {
			fetched, err := sessions.Fetch(ctx, Lookup{
				Key: created[0].Key,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("Owner", func(t *testing.T) {
			fetched, err := sessions.Fetch(ctx, Lookup{
				Owner: created[0].Owner,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
	})
}

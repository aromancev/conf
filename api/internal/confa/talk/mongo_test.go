package talk

import (
	"context"

	"github.com/aromancev/confa/internal/confa"

	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMongo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type dependencies struct {
		talks *Mongo
	}
	setup := func() dependencies {
		db := dockerMongo(t)
		return dependencies{
			talks: NewMongo(db),
		}
	}

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			deps := setup()

			request := Talk{
				ID:      uuid.New(),
				Owner:   uuid.New(),
				Confa:   uuid.New(),
				Speaker: uuid.New(),
				Handle:  "test",
			}

			created, err := deps.talks.Create(ctx, request)
			require.NoError(t, err)

			fetched, err := deps.talks.Fetch(ctx, Lookup{
				ID: request.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
		t.Run("UUID handle", func(t *testing.T) {
			t.Parallel()

			deps := setup()

			request := Talk{
				ID:      uuid.New(),
				Owner:   uuid.New(),
				Confa:   uuid.New(),
				Speaker: uuid.New(),
				Handle:  uuid.New().String(),
			}

			created, err := deps.talks.Create(ctx, request)
			require.NoError(t, err)

			fetched, err := deps.talks.Fetch(ctx, Lookup{
				ID: request.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
		t.Run("Duplicated Entry", func(t *testing.T) {
			t.Parallel()

			deps := setup()

			_, err := deps.talks.Create(ctx, Talk{
				ID:      uuid.New(),
				Owner:   uuid.New(),
				Confa:   uuid.New(),
				Speaker: uuid.New(),
				Handle:  "test",
			})
			require.NoError(t, err)
			_, err = deps.talks.Create(ctx, Talk{
				ID:      uuid.New(),
				Owner:   uuid.New(),
				Confa:   uuid.New(),
				Speaker: uuid.New(),
				Handle:  "test",
			})
			assert.ErrorIs(t, err, ErrDuplicateEntry)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Skip()
		t.Parallel()

		db := dockerMongo(t)
		talks := NewMongo(db)
		confas := confa.NewMongo(db)

		conf := confa.Confa{
			ID:     uuid.New(),
			Owner:  uuid.New(),
			Handle: "test",
		}

		tlk := Talk{
			ID:     uuid.New(),
			Owner:  uuid.New(),
			Confa:  conf.ID,
			Handle: "test",
		}

		_, err := confas.Create(ctx, conf)
		require.NoError(t, err)

		createdTalk, err := talks.Create(ctx, tlk)
		require.NoError(t, err)

		t.Run("ID", func(t *testing.T) {
			fetchedTalk, err := talks.Fetch(ctx, Lookup{
				ID: tlk.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, createdTalk, fetchedTalk)
		})

		t.Run("Confa", func(t *testing.T) {
			fetchedTalk, err := talks.Fetch(ctx, Lookup{
				Confa: conf.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, createdTalk, fetchedTalk)
		})
	})
}

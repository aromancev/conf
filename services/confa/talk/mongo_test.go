package talk

import (
	"context"

	"github.com/aromancev/confa/confa"

	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMongo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			talks := NewMongo(dockerMongo(t))

			request := Talk{
				ID:      uuid.New(),
				Owner:   uuid.New(),
				Confa:   uuid.New(),
				Speaker: uuid.New(),
				Room:    uuid.New(),
				State:   StateLive,
				Handle:  "test",
			}

			created, err := talks.Create(ctx, request)
			require.NoError(t, err)

			fetched, err := talks.Fetch(ctx, Lookup{
				ID: request.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
		t.Run("UUID handle", func(t *testing.T) {
			t.Parallel()

			talks := NewMongo(dockerMongo(t))

			request := Talk{
				ID:      uuid.New(),
				Owner:   uuid.New(),
				Confa:   uuid.New(),
				Speaker: uuid.New(),
				Room:    uuid.New(),
				State:   StateLive,
				Handle:  uuid.New().String(),
			}

			created, err := talks.Create(ctx, request)
			require.NoError(t, err)

			fetched, err := talks.Fetch(ctx, Lookup{
				ID: request.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
		t.Run("Duplicated Entry", func(t *testing.T) {
			t.Parallel()

			talks := NewMongo(dockerMongo(t))

			_, err := talks.Create(ctx, Talk{
				ID:      uuid.New(),
				Owner:   uuid.New(),
				Confa:   uuid.New(),
				Speaker: uuid.New(),
				Room:    uuid.New(),
				State:   StateLive,
				Handle:  "test",
			})
			require.NoError(t, err)
			_, err = talks.Create(ctx, Talk{
				ID:      uuid.New(),
				Owner:   uuid.New(),
				Confa:   uuid.New(),
				Speaker: uuid.New(),
				Room:    uuid.New(),
				State:   StateLive,
				Handle:  "test",
			})
			assert.ErrorIs(t, err, ErrDuplicateEntry)
		})
	})

	t.Run("UpdateOne", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			ctx := context.Background()

			confas := NewMongo(dockerMongo(t))

			request := Talk{
				ID:      uuid.New(),
				Owner:   uuid.New(),
				Confa:   uuid.New(),
				Speaker: uuid.New(),
				Room:    uuid.New(),
				State:   StateLive,
				Handle:  "1111",
			}
			created, err := confas.Create(ctx, request)
			require.NoError(t, err)

			request = created[0]
			request.Handle = "2222"
			request.Title = "title"
			updated, err := confas.UpdateOne(ctx, Lookup{ID: request.ID}, Update{Handle: &request.Handle, Title: &request.Title})
			require.NoError(t, err)
			require.EqualValues(t, request, updated)

			fetched, err := confas.FetchOne(ctx, Lookup{
				ID: request.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, updated, fetched)
		})

		t.Run("Not found returns correct error", func(t *testing.T) {
			ctx := context.Background()

			confas := NewMongo(dockerMongo(t))

			handle := "test"
			_, err := confas.UpdateOne(ctx, Lookup{ID: uuid.New()}, Update{Handle: &handle})
			require.ErrorIs(t, err, ErrNotFound)
		})

		t.Run("Duplicated entry returns correct error", func(t *testing.T) {
			ctx := context.Background()

			confas := NewMongo(dockerMongo(t))

			created, err := confas.Create(
				ctx,
				Talk{
					ID:      uuid.New(),
					Owner:   uuid.New(),
					Confa:   uuid.New(),
					Speaker: uuid.New(),
					Room:    uuid.New(),
					State:   StateLive,
					Handle:  uuid.NewString(),
				},
				Talk{
					ID:      uuid.New(),
					Owner:   uuid.New(),
					Confa:   uuid.New(),
					Speaker: uuid.New(),
					Room:    uuid.New(),
					State:   StateLive,
					Handle:  uuid.NewString(),
				},
			)
			require.NoError(t, err)

			_, err = confas.UpdateOne(ctx, Lookup{ID: created[0].ID}, Update{Handle: &created[1].Handle})
			require.ErrorIs(t, err, ErrDuplicateEntry)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		db := dockerMongo(t)
		talks := NewMongo(db)
		confas := confa.NewMongo(db)

		conf := confa.Confa{
			ID:     uuid.New(),
			Owner:  uuid.New(),
			Handle: "test",
		}
		_, err := confas.Create(ctx, conf)
		require.NoError(t, err)

		created, err := talks.Create(ctx,
			Talk{
				ID:      uuid.UUID{1},
				Owner:   uuid.New(),
				Speaker: uuid.New(),
				Room:    uuid.New(),
				State:   StateLive,
				Confa:   conf.ID,
				Handle:  "test1",
			},
			Talk{
				ID:      uuid.UUID{2},
				Owner:   uuid.New(),
				Speaker: uuid.New(),
				Room:    uuid.New(),
				State:   StateEnded,
				Confa:   conf.ID,
				Handle:  "test2",
			},
			Talk{
				ID:      uuid.UUID{3},
				Owner:   uuid.New(),
				Speaker: uuid.New(),
				Room:    uuid.New(),
				State:   StateRecording,
				Confa:   uuid.New(),
				Handle:  "test3",
			},
		)
		require.NoError(t, err)

		t.Run("By ID", func(t *testing.T) {
			fetched, err := talks.Fetch(ctx, Lookup{
				ID: created[0].ID,
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, created[:1], fetched)
		})

		t.Run("By Confa", func(t *testing.T) {
			fetched, err := talks.Fetch(ctx, Lookup{
				Confa: conf.ID,
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, created[:2], fetched)
		})

		t.Run("By State In", func(t *testing.T) {
			fetched, err := talks.Fetch(ctx, Lookup{
				StateIn: []State{StateEnded},
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, []Talk{created[1]}, fetched)
		})

		t.Run("Pagination works in both directions", func(t *testing.T) {
			fetched, err := talks.Fetch(ctx, Lookup{
				From: From{
					ID: created[1].ID,
				},
				Limit: 1,
				Asc:   true,
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, []Talk{created[2]}, fetched)

			fetched, err = talks.Fetch(ctx, Lookup{
				From: From{
					ID: created[1].ID,
				},
				Limit: 1,
				Asc:   false,
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, []Talk{created[0]}, fetched)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()

		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			talks := NewMongo(dockerMongo(t))

			request := Talk{
				ID:      uuid.New(),
				Owner:   uuid.New(),
				Confa:   uuid.New(),
				Speaker: uuid.New(),
				Room:    uuid.New(),
				State:   StateLive,
				Handle:  "test",
			}

			_, err := talks.Create(ctx, request)
			require.NoError(t, err)

			res, err := talks.Delete(ctx, Lookup{
				ID: request.ID,
			})
			require.NoError(t, err)
			assert.EqualValues(t, 1, res.Updated)

			_, err = talks.FetchOne(ctx, Lookup{
				ID: request.ID,
			})
			assert.ErrorIs(t, err, ErrNotFound)
		})
	})
}

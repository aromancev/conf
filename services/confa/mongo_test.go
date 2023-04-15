package confa

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

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		t.Run("String handle", func(t *testing.T) {
			t.Parallel()

			confas := NewMongo(dockerMongo(t))

			request := Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test1",
			}
			created, err := confas.Create(ctx, request)
			require.NoError(t, err)

			fetched, err := confas.Fetch(ctx, Lookup{
				ID:    request.ID,
				Owner: request.Owner,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
		t.Run("UUID handle", func(t *testing.T) {
			t.Parallel()

			confas := NewMongo(dockerMongo(t))

			request := Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: uuid.New().String(),
			}
			_, err := confas.Create(ctx, request)
			require.NoError(t, err)
		})
		t.Run("Duplicated entry returns correct error", func(t *testing.T) {
			t.Parallel()

			confas := NewMongo(dockerMongo(t))

			_, err := confas.Create(ctx, Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test",
			})
			require.NoError(t, err)
			_, err = confas.Create(ctx, Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test",
			})
			require.ErrorIs(t, err, ErrDuplicateEntry)

			_, err = confas.Create(
				ctx,
				Confa{
					ID:     uuid.New(),
					Owner:  uuid.New(),
					Handle: "test2",
				},
				Confa{
					ID:     uuid.New(),
					Owner:  uuid.New(),
					Handle: "test2",
				},
			)
			require.ErrorIs(t, err, ErrDuplicateEntry)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			ctx := context.Background()

			confas := NewMongo(dockerMongo(t))

			request := Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "1111",
			}
			created, err := confas.Create(ctx, request)
			require.NoError(t, err)

			updated := created[0]
			updated.Handle = "2222"
			updated.Title = "title"
			res, err := confas.Update(ctx, Lookup{ID: updated.ID}, Update{Handle: &updated.Handle, Title: &updated.Title})
			require.NoError(t, err)
			require.EqualValues(t, 1, res.Updated)

			fetched, err := confas.FetchOne(ctx, Lookup{
				ID: request.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, updated, fetched)
		})

		t.Run("Duplicated entry returns correct error", func(t *testing.T) {
			ctx := context.Background()

			confas := NewMongo(dockerMongo(t))

			created, err := confas.Create(
				ctx,
				Confa{
					ID:     uuid.New(),
					Owner:  uuid.New(),
					Handle: uuid.NewString(),
				},
				Confa{
					ID:     uuid.New(),
					Owner:  uuid.New(),
					Handle: uuid.NewString(),
				},
			)
			require.NoError(t, err)

			res, err := confas.Update(ctx, Lookup{ID: created[0].ID}, Update{Handle: &created[1].Handle})
			require.ErrorIs(t, err, ErrDuplicateEntry)
			require.EqualValues(t, 0, res.Updated)
		})
	})

	t.Run("UpdateOne", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			ctx := context.Background()

			confas := NewMongo(dockerMongo(t))

			request := Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "1111",
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
				Confa{
					ID:     uuid.New(),
					Owner:  uuid.New(),
					Handle: uuid.NewString(),
				},
				Confa{
					ID:     uuid.New(),
					Owner:  uuid.New(),
					Handle: uuid.NewString(),
				},
			)
			require.NoError(t, err)

			_, err = confas.UpdateOne(ctx, Lookup{ID: created[0].ID}, Update{Handle: &created[1].Handle})
			require.ErrorIs(t, err, ErrDuplicateEntry)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		confas := NewMongo(dockerMongo(t))
		created, err := confas.Create(
			ctx,
			Confa{
				ID:     uuid.UUID{1},
				Owner:  uuid.New(),
				Handle: "test",
			},
			Confa{
				ID:     uuid.UUID{2},
				Owner:  uuid.New(),
				Handle: uuid.NewString(),
			},
			Confa{
				ID:     uuid.UUID{3},
				Owner:  uuid.New(),
				Handle: uuid.NewString(),
			},
		)
		require.NoError(t, err)

		t.Run("By id", func(t *testing.T) {
			fetched, err := confas.Fetch(ctx, Lookup{
				ID: created[1].ID,
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, []Confa{created[1]}, fetched)
		})

		t.Run("By owner", func(t *testing.T) {
			fetched, err := confas.Fetch(ctx, Lookup{
				Owner: created[1].Owner,
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, []Confa{created[1]}, fetched)
		})

		t.Run("Pagination works in both directions", func(t *testing.T) {
			fetched, err := confas.Fetch(ctx, Lookup{
				From: From{
					ID: created[1].ID,
				},
				Limit: 1,
				Asc:   true,
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, []Confa{created[2]}, fetched)

			fetched, err = confas.Fetch(ctx, Lookup{
				From: From{
					ID: created[1].ID,
				},
				Limit: 1,
				Asc:   false,
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, []Confa{created[0]}, fetched)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()

		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			confas := NewMongo(dockerMongo(t))

			request := Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test1",
			}
			_, err := confas.Create(ctx, request)
			require.NoError(t, err)

			res, err := confas.Delete(ctx, Lookup{
				ID: request.ID,
			})
			require.NoError(t, err)
			assert.EqualValues(t, 1, res.Updated)

			_, err = confas.FetchOne(ctx, Lookup{
				ID: request.ID,
			})
			assert.ErrorIs(t, err, ErrNotFound)
		})
	})
}

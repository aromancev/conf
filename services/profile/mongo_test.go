package profile

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

		t.Run("Hapy", func(t *testing.T) {
			t.Parallel()

			profiles := NewMongo(dockerMongo(t))

			request := Profile{
				ID:        uuid.New(),
				Owner:     uuid.New(),
				Handle:    "test",
				GivenName: "test",
			}
			created, err := profiles.Create(ctx, request)
			require.NoError(t, err)
			assert.NotZero(t, created[0].CreatedAt)

			fetched, err := profiles.Fetch(ctx, Lookup{
				ID: request.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("Owner duplicates not allowed", func(t *testing.T) {
			t.Parallel()

			profiles := NewMongo(dockerMongo(t))

			_, err := profiles.Create(
				ctx,
				Profile{
					ID:         uuid.New(),
					Owner:      uuid.UUID{1},
					Handle:     "test-1",
					GivenName:  "Rick",
					FamilyName: "Sanchez",
				},
				Profile{
					ID:         uuid.New(),
					Owner:      uuid.UUID{1},
					Handle:     "test-2",
					GivenName:  "Rick",
					FamilyName: "Sanchez",
				},
			)
			require.ErrorIs(t, err, ErrDuplicateEntry)
		})

		t.Run("Handle duplicates not allowed", func(t *testing.T) {
			t.Parallel()

			profiles := NewMongo(dockerMongo(t))

			_, err := profiles.Create(
				ctx,
				Profile{
					ID:         uuid.New(),
					Owner:      uuid.New(),
					Handle:     "test",
					GivenName:  "Rick",
					FamilyName: "Sanchez",
				},
				Profile{
					ID:         uuid.New(),
					Owner:      uuid.New(),
					Handle:     "test",
					GivenName:  "Rick",
					FamilyName: "Sanchez",
				},
			)
			require.ErrorIs(t, err, ErrDuplicateEntry)
		})
	})

	t.Run("CreateOrUpdate", func(t *testing.T) {
		t.Parallel()

		t.Run("Happy", func(t *testing.T) {
			t.Parallel()

			profiles := NewMongo(dockerMongo(t))

			request := Profile{
				ID:        uuid.New(),
				Owner:     uuid.New(),
				Handle:    "test",
				GivenName: "test",
			}
			created, err := profiles.CreateOrUpdate(ctx, request)
			require.NoError(t, err)

			fetched, err := profiles.Fetch(ctx, Lookup{
				Owners: []uuid.UUID{request.Owner},
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched[0])
		})

		t.Run("Updates fields", func(t *testing.T) {
			t.Parallel()

			profiles := NewMongo(dockerMongo(t))

			request := Profile{
				ID:        uuid.New(),
				Owner:     uuid.New(),
				Handle:    "test",
				GivenName: "test",
			}
			created, err := profiles.CreateOrUpdate(ctx, request)
			require.NoError(t, err)

			created.GivenName = "changed"
			created.AvatarThumbnail = Image{
				Format: "jpeg",
				Data:   []byte{1},
			}
			updated, err := profiles.CreateOrUpdate(ctx, created)
			require.NoError(t, err)
			assert.Equal(t, created, updated)
		})

		t.Run("Does not override create only fields", func(t *testing.T) {
			t.Parallel()

			profiles := NewMongo(dockerMongo(t))

			request := Profile{
				ID:        uuid.New(),
				Owner:     uuid.New(),
				Handle:    "test",
				GivenName: "test",
			}
			created, err := profiles.CreateOrUpdate(ctx, request)
			require.NoError(t, err)

			updated, err := profiles.CreateOrUpdate(ctx, Profile{
				ID:        uuid.New(),
				Owner:     request.Owner,
				GivenName: "changed",
			})
			created.GivenName = "changed"
			require.NoError(t, err)
			assert.Equal(t, created, updated)
		})

		t.Run("Fills handle if empty", func(t *testing.T) {
			profiles := NewMongo(dockerMongo(t))

			request := Profile{
				ID:        uuid.New(),
				Owner:     uuid.New(),
				GivenName: "test",
			}
			created, err := profiles.CreateOrUpdate(ctx, request)
			require.NoError(t, err)
			assert.Equal(t, request.ID.String(), created.Handle)
		})

		t.Run("Empty fields do not override", func(t *testing.T) {
			profiles := NewMongo(dockerMongo(t))

			request := Profile{
				ID:        uuid.New(),
				Owner:     uuid.New(),
				Handle:    "test",
				GivenName: "test",
			}
			_, err := profiles.CreateOrUpdate(ctx, request)
			require.NoError(t, err)

			request.GivenName = ""
			created, err := profiles.CreateOrUpdate(ctx, request)
			require.NoError(t, err)
			require.Equal(t, "test", created.GivenName)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		profiles := NewMongo(dockerMongo(t))

		created, err := profiles.Create(
			ctx,
			Profile{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test-1",
			},
			Profile{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test-2",
			},
			Profile{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test-3",
			},
		)
		require.NoError(t, err)

		t.Run("by handle", func(t *testing.T) {
			fetched, err := profiles.Fetch(ctx, Lookup{
				Handle: created[0].Handle,
			})
			require.NoError(t, err)
			assert.Equal(t, []Profile{created[0]}, fetched)
		})

		t.Run("by owner", func(t *testing.T) {
			fetched, err := profiles.Fetch(ctx, Lookup{
				Owners: []uuid.UUID{created[0].Owner},
			})
			require.NoError(t, err)
			assert.Equal(t, created[0], fetched[0])
		})

		t.Run("with limit and offset", func(t *testing.T) {
			fetched, err := profiles.Fetch(ctx, Lookup{
				Limit: 1,
			})
			require.NoError(t, err)
			assert.Equal(t, 1, len(fetched))

			// 3 in total, skipped one.
			fetched, err = profiles.Fetch(ctx, Lookup{
				From: Cursor{ID: fetched[0].ID},
			})
			require.NoError(t, err)
			assert.Equal(t, 2, len(fetched))
		})
	})
}

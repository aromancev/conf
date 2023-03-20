package user

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

		tt := [...]struct {
			name             string
			create, existing []Ident
			err              error
		}{
			{
				name: "Happy path",
				create: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test@confa.io",
					},
				},
				err: nil,
			},
			{
				name: "Same platform is allowed",
				existing: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test1@confa.io",
					},
				},
				create: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test2@confa.io",
					},
				},
				err: nil,
			},
			{
				name: "Same value is allowed",
				existing: []Ident{
					{
						Platform: PlatformGithub,
						Value:    "test",
					},
				},
				create: []Ident{
					{
						Platform: PlatformTwitter,
						Value:    "test",
					},
				},
				err: nil,
			},
			{
				name: "Same platform and value is not allowed",
				existing: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test@confa.io",
					},
				},
				create: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test@confa.io",
					},
				},
				err: ErrDuplicateEntry,
			},
			{
				name: "Same platform in single user is allowed",
				create: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test1@confa.io",
					},
					{
						Platform: PlatformEmail,
						Value:    "test2@confa.io",
					},
				},
				err: nil,
			},
			{
				name: "Same value in single user is allowed",
				create: []Ident{
					{
						Platform: PlatformTwitter,
						Value:    "test",
					},
					{
						Platform: PlatformGithub,
						Value:    "test",
					},
				},
				err: nil,
			},
			{
				name: "Same platform and value in single user is not allowed",
				create: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test@confa.io",
					},
					{
						Platform: PlatformEmail,
						Value:    "test@confa.io",
					},
				},
				err: ErrValidation,
			},
			{
				name: "Value is normalized",
				create: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "Test@confa.io",
					},
					{
						Platform: PlatformEmail,
						Value:    "test@confa.io",
					},
				},
				err: ErrValidation,
			},
		}

		for _, c := range tt {
			c := c // Parallel execution protection.

			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				users := NewMongo(dockerMongo(t))
				if len(c.existing) != 0 {
					_, err := users.Create(ctx, User{
						ID:     uuid.New(),
						Idents: c.existing,
					})
					require.NoError(t, err)
				}

				created, err := users.Create(ctx, User{
					ID:     uuid.New(),
					Idents: c.create,
				})
				assert.ErrorIs(t, err, c.err)
				if err != nil {
					return
				}
				assert.NotZero(t, created[0].CreatedAt)

				fetched, err := users.Fetch(ctx, Lookup{
					ID: created[0].ID,
				})
				require.NoError(t, err)
				assert.Equal(t, created, fetched)
			})
		}
	})

	t.Run("GetOrCreate", func(t *testing.T) {
		t.Parallel()

		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			// First time just creates a new User.
			created, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test@confa.io",
					},
				},
			})
			require.NoError(t, err)
			assert.NotZero(t, created.CreatedAt)

			// Calling with the same Identifier returns the same User.
			fetched, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test@confa.io",
					},
				},
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("Matching many identifiers but does not modify the original", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			// First time just creates a new User.
			created, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test1@confa.io",
					},
				},
			})
			require.NoError(t, err)
			assert.NotZero(t, created.CreatedAt)

			// Calling with many Identifiers returns the same User.
			fetched, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test1@confa.io",
					},
					{
						Platform: PlatformGithub,
						Value:    "test",
					},
				},
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))

			created, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test@confa.io",
					},
				},
			})
			require.NoError(t, err)

			res, err := users.Update(ctx, Lookup{ID: created.ID}, Update{
				PasswordHash: []byte{1},
			})
			require.NoError(t, err)
			assert.EqualValues(t, 1, res.Updated)

			fetched, err := users.FetchOne(ctx, Lookup{ID: created.ID})
			require.NoError(t, err)
			assert.EqualValues(t, []byte{1}, fetched.PasswordHash)
		})

		t.Run("Filter by ident", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))

			ident := Ident{
				Platform: PlatformEmail,
				Value:    "test@confa.io",
			}
			_, err := users.GetOrCreate(ctx, User{
				ID:     uuid.New(),
				Idents: []Ident{ident},
			})
			require.NoError(t, err)

			res, err := users.Update(
				ctx,
				Lookup{
					Idents: []Ident{
						{
							Platform: ident.Platform,
							Value:    ident.Value,
						},
					},
				},
				Update{
					PasswordHash: []byte{1},
				},
			)
			require.NoError(t, err)
			assert.EqualValues(t, 1, res.Updated)
		})

		t.Run("Without password", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))

			ident := Ident{
				Platform: PlatformGithub,
				Value:    "test",
			}
			_, err := users.GetOrCreate(ctx, User{
				ID:           uuid.New(),
				Idents:       []Ident{ident},
				PasswordHash: []byte{1},
			})
			require.NoError(t, err)

			res, err := users.Update(
				ctx,
				Lookup{
					WithoutPassword: true,
				},
				Update{
					PasswordHash: []byte{1},
				},
			)
			require.NoError(t, err)
			assert.EqualValues(t, 0, res.Updated)
		})
	})

	t.Run("UpdateOne", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))

			ident := Ident{
				Platform: PlatformGithub,
				Value:    "test",
			}
			created, err := users.GetOrCreate(ctx, User{
				ID:     uuid.New(),
				Idents: []Ident{ident},
			})
			require.NoError(t, err)

			created.PasswordHash = []byte{1}
			updated, err := users.UpdateOne(
				ctx,
				Lookup{
					Idents: []Ident{
						{
							Platform: ident.Platform,
							Value:    ident.Value,
						},
					},
				},
				Update{
					PasswordHash: []byte{1},
				},
			)
			require.NoError(t, err)
			assert.EqualValues(t, created, updated)
		})
		t.Run("Non existend returns not found", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))

			_, err := users.UpdateOne(
				ctx,
				Lookup{
					Idents: []Ident{
						{
							Platform: PlatformGithub,
							Value:    "any",
						},
					},
				},
				Update{
					PasswordHash: []byte{1},
				},
			)
			require.ErrorIs(t, err, ErrNotFound)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			// First time just creates a new User.
			created, err := users.Create(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test@confa.io",
					},
				},
			})
			require.NoError(t, err)

			// Calling with the same Identifier returns the same User.
			fetched, err := users.Fetch(ctx, Lookup{
				ID: created[0].ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("Normalizes identifiers", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			// First time just creates a new User.
			created, err := users.Create(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "test@confa.io",
					},
				},
			})
			require.NoError(t, err)

			// Calling with the same Identifier returns the same User.
			fetched, err := users.Fetch(ctx, Lookup{
				Idents: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "Test@confa.io",
					},
				},
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
	})
}

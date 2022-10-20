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
						Value:    "2",
					},
				},
				err: nil,
			},
			{
				name: "Same platform is allowed",
				existing: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "1",
					},
				},
				create: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "2",
					},
				},
				err: nil,
			},
			{
				name: "Same value is allowed",
				existing: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "1",
					},
				},
				create: []Ident{
					{
						Platform: PlatformTwitter,
						Value:    "1",
					},
				},
				err: nil,
			},
			{
				name: "Same platform and value is not allowed",
				existing: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "1",
					},
				},
				create: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "1",
					},
				},
				err: ErrDuplicatedEntry,
			},
			{
				name: "Same platform in single user is allowed",
				create: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "1",
					},
					{
						Platform: PlatformEmail,
						Value:    "2",
					},
				},
				err: nil,
			},
			{
				name: "Same value in single user is allowed",
				create: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "1",
					},
					{
						Platform: PlatformGithub,
						Value:    "1",
					},
				},
				err: nil,
			},
			{
				name: "Same platform and value in single user is not allowed",
				create: []Ident{
					{
						Platform: PlatformEmail,
						Value:    "1",
					},
					{
						Platform: PlatformEmail,
						Value:    "1",
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
					for i := range c.existing {
						c.existing[i].ID = uuid.New()
					}
					_, err := users.Create(ctx, User{
						ID:     uuid.New(),
						Idents: c.existing,
					})
					require.NoError(t, err)
				}

				for i := range c.create {
					c.create[i].ID = uuid.New()
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
				for _, ident := range created[0].Idents {
					assert.NotZero(t, ident.CreatedAt)
				}

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
						ID:       uuid.New(),
						Platform: PlatformEmail,
						Value:    "1",
					},
				},
			})
			require.NoError(t, err)
			assert.NotZero(t, created.CreatedAt)
			assert.NotZero(t, created.Idents[0].CreatedAt)

			// Calling with the same Identifier returns the same User.
			fetched, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						ID:       uuid.New(),
						Platform: PlatformEmail,
						Value:    "1",
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
						ID:       uuid.New(),
						Platform: PlatformEmail,
						Value:    "1",
					},
				},
			})
			require.NoError(t, err)
			assert.NotZero(t, created.CreatedAt)
			assert.NotZero(t, created.Idents[0].CreatedAt)

			// Calling with many Identifiers returns the same User.
			fetched, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						ID:       uuid.New(),
						Platform: PlatformEmail,
						Value:    "1",
					},
					{
						ID:       uuid.New(),
						Platform: PlatformGithub,
						Value:    "2",
					},
				},
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
	})
}

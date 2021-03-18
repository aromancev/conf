package iam

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestIdentSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Run("Create duplicate returns correct error", func(t *testing.T) {
			t.Parallel()

			pg, done := double.NewDocker("", migrate)
			defer done()

			users := NewUserSQL()
			idents := NewIdentSQL()

			user := User{ID: uuid.New()}
			_, _ = users.Create(ctx, pg, user)

			ident := Ident{
				ID:       uuid.New(),
				Owner:    user.ID,
				Platform: PlatformEmail,
				Value:    uuid.NewString(),
			}

			_, err := idents.Create(ctx, pg, ident)
			require.NoError(t, err)

			ident.ID = uuid.New()
			_, err = idents.Create(ctx, pg, ident)
			require.True(t, err == ErrDuplicatedEntry)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()

		users := NewUserSQL()
		idents := NewIdentSQL()

		user := User{ID: uuid.New()}
		_, _ = users.Create(ctx, pg, user)

		ident := Ident{
			ID:       uuid.New(),
			Owner:    user.ID,
			Platform: PlatformEmail,
			Value:    uuid.NewString(),
		}
		created, err := idents.Create(ctx, pg, ident)
		require.NoError(t, err)

		t.Run("ID", func(t *testing.T) {
			fetched, err := idents.Fetch(ctx, pg, IdentLookup{
				ID: ident.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("Owner", func(t *testing.T) {
			fetched, err := idents.Fetch(ctx, pg, IdentLookup{
				Owner: user.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("Matching", func(t *testing.T) {
			fetched, err := idents.Fetch(ctx, pg, IdentLookup{
				Matching: []Ident{
					{Platform: ident.Platform, Value: ident.Value},
				}},
			)
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
	})
}

func TestUserSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()

		users := NewUserSQL()

		user := User{ID: uuid.New()}
		createdUser, _ := users.Create(ctx, pg, user)

		t.Run("Fetch", func(t *testing.T) {
			fetchedUser, err := users.Fetch(ctx, pg, UserLookup{
				ID: user.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, createdUser, fetchedUser)
		})

		t.Run("FetchOne", func(t *testing.T) {
			fetchedUser, err := users.FetchOne(ctx, pg, UserLookup{
				ID: user.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, createdUser[0], fetchedUser)
		})

	})
}

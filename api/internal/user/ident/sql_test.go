package ident

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
	"github.com/aromancev/confa/internal/user"
)

func TestSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Run("Create duplicate returns correct error", func(t *testing.T) {
			t.Parallel()

			pg, done := double.NewDocker("", migrate)
			defer done()

			users := user.NewSQL()
			idents := NewSQL()

			usr := user.User{ID: uuid.New()}
			_, _ = users.Create(ctx, pg, usr)

			ident := Ident{
				ID:       uuid.New(),
				Owner:    usr.ID,
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

		users := user.NewSQL()
		idents := NewSQL()

		usr := user.User{ID: uuid.New()}
		_, _ = users.Create(ctx, pg, usr)

		ident := Ident{
			ID:       uuid.New(),
			Owner:    usr.ID,
			Platform: PlatformEmail,
			Value:    uuid.NewString(),
		}
		created, err := idents.Create(ctx, pg, ident)
		require.NoError(t, err)

		t.Run("ID", func(t *testing.T) {
			fetched, err := idents.Fetch(ctx, pg, Lookup{
				ID: ident.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("Owner", func(t *testing.T) {
			fetched, err := idents.Fetch(ctx, pg, Lookup{
				Owner: usr.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("Matching", func(t *testing.T) {
			fetched, err := idents.Fetch(ctx, pg, Lookup{
				Matching: []Ident{
					{Platform: ident.Platform, Value: ident.Value},
				}},
			)
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
	})
}

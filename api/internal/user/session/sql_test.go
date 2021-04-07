package session

import (
	"context"
	"testing"

	"github.com/aromancev/confa/internal/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()
			pg, done := double.NewDocker("", migrate)
			defer done()

			users := user.NewSQL()
			sessions := NewSQL()

			usr := user.User{
				ID: uuid.New(),
			}
			sess := NewSession()
			sess.Owner = usr.ID

			_, err := users.Create(ctx, pg, usr)
			require.NoError(t, err)

			createdSession, err := sessions.Create(ctx, pg, sess)
			require.NoError(t, err)

			fetchedSession, err := sessions.Fetch(ctx, pg, Lookup{
				Key:   sess.Key,
				Owner: sess.Owner,
			})
			require.NoError(t, err)
			assert.Equal(t, createdSession, fetchedSession)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()

		users := user.NewSQL()
		sessions := NewSQL()

		usr := user.User{
			ID: uuid.New(),
		}
		sess := NewSession()
		sess.Owner = usr.ID

		_, err := users.Create(ctx, pg, usr)
		require.NoError(t, err)

		createdSession, err := sessions.Create(ctx, pg, sess)
		require.NoError(t, err)

		t.Run("Key", func(t *testing.T) {
			fetchedSession, err := sessions.Fetch(ctx, pg, Lookup{
				Key: sess.Key,
			})
			require.NoError(t, err)
			assert.Equal(t, createdSession, fetchedSession)
		})

		t.Run("Owner", func(t *testing.T) {
			fetchedSession, err := sessions.Fetch(ctx, pg, Lookup{
				Owner: sess.Owner,
			})
			require.NoError(t, err)
			assert.Equal(t, createdSession, fetchedSession)
		})
	})
}

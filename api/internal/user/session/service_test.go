package session

import (
	"context"
	"testing"

	"github.com/aromancev/confa/internal/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()

		users := user.NewSQL()
		sessions := NewSQL()
		sessionCRUD := NewCRUD(pg, sessions)

		usr := user.User{
			ID: uuid.New(),
		}

		_, err := users.Create(ctx, pg, usr)
		require.NoError(t, err)

		createdSession, err := sessionCRUD.Create(ctx, usr.ID)
		require.NoError(t, err)

		fetchedSession, err := sessions.FetchOne(ctx, pg, Lookup{Key: createdSession.Key})
		require.NoError(t, err)
		require.Equal(t, createdSession, fetchedSession)
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()

		users := user.NewSQL()
		sessions := NewSQL()
		sessionCRUD := NewCRUD(pg, sessions)

		usr := user.User{
			ID: uuid.New(),
		}

		_, err := users.Create(ctx, pg, usr)
		require.NoError(t, err)

		sess := NewSession()
		sess.Owner = usr.ID
		created, err := sessions.Create(ctx, pg, sess)
		require.NoError(t, err)

		createdSession := created[0]
		fetchedSession, err := sessionCRUD.Fetch(ctx, createdSession.Key)
		require.NoError(t, err)
		require.Equal(t, createdSession, fetchedSession)
	})
}

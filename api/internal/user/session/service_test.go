package session

import (
	"context"
	"github.com/aromancev/confa/internal/user"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()

		userCRUD := user.NewCRUD(pg, user.NewSQL())
		sessionCRUD := NewCRUD(pg, NewSQL())

		usr := user.User{}
		sess := Session{
			Owner: uuid.New(),
		}

		usr, err := userCRUD.Create(ctx, usr)
		require.NoError(t, err)

		createdSession, err := sessionCRUD.Create(ctx, usr.ID, sess)
		require.NoError(t, err)

		fetchedSession, err := sessionCRUD.Fetch(ctx, createdSession.Key)
		require.NoError(t, err)
		require.Equal(t, createdSession, fetchedSession)
	})

}

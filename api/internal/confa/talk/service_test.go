package talk

import (
	"context"
	"github.com/aromancev/confa/internal/confa"
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
		confaCRUD := confa.NewCRUD(pg, confa.NewSQL())
		talkCRUD := NewCRUD(pg, NewSQL(), confaCRUD)

		userID := uuid.New()
		requestConfa := confa.Confa{
			Handle: "test",
		}
		requestTalk := Talk{
			Handle:  "test",
			Speaker: userID,
		}

		createdConfa, err := confaCRUD.Create(ctx, userID, requestConfa)
		require.NoError(t, err)

		createdTalk, err := talkCRUD.Create(ctx, createdConfa.ID, userID, requestTalk)
		require.NoError(t, err)

		fetchedTalk, err := talkCRUD.Fetch(ctx, createdTalk.ID)
		require.NoError(t, err)

		require.Equal(t, createdTalk, fetchedTalk)
	})

	t.Run("Permission denied", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()
		confaCRUD := confa.NewCRUD(pg, confa.NewSQL())
		talkCRUD := NewCRUD(pg, NewSQL(), confaCRUD)

		userID := uuid.New()
		wronguserID := uuid.New()
		requestConfa := confa.Confa{
			Handle: "test",
		}
		requestTalk := Talk{
			Handle:  "test",
			Speaker: userID,
		}

		createdConfa, err := confaCRUD.Create(ctx, userID, requestConfa)
		require.NoError(t, err)

		_, err = talkCRUD.Create(ctx, createdConfa.ID, wronguserID, requestTalk)
		require.ErrorIs(t, err, ErrPermissionDenied)
	})
}

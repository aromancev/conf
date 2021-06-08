package clap

import (
	"context"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/confa/talk"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("FetchByTalk", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()
		clapCRUD := NewCRUD(pg, NewSQL())
		confaCRUD := confa.NewCRUD(pg, confa.NewSQL())
		talkCRUD := talk.NewCRUD(pg, talk.NewSQL(), confaCRUD)

		userID := uuid.New()
		requestConfa := confa.Confa{
			Handle: "test",
		}
		requestTalk := talk.Talk{
			Handle:  "test",
			Speaker: userID,
		}

		createdConfa, err := confaCRUD.Create(ctx, userID, requestConfa)
		require.NoError(t, err)

		createdTalk, err := talkCRUD.Create(ctx, createdConfa.ID, userID, requestTalk)
		require.NoError(t, err)

		requestClap := Clap{
			Confa: createdConfa.ID,
			Talk:  createdTalk.ID,
			Claps: 2,
		}
		err = clapCRUD.CreateOrUpdate(ctx, userID, requestClap)
		require.NoError(t, err)
		lookup := Lookup{
			Talk: createdTalk.ID,
		}
		claps, err := clapCRUD.Aggregate(ctx, lookup)
		require.NoError(t, err)
		require.Equal(t, claps, 2)
	})

	t.Run("FetchByConfa", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()
		clapCRUD := NewCRUD(pg, NewSQL())
		confaCRUD := confa.NewCRUD(pg, confa.NewSQL())
		talkCRUD := talk.NewCRUD(pg, talk.NewSQL(), confaCRUD)

		userID := uuid.New()
		requestConfa := confa.Confa{
			Handle: "test",
		}
		requestTalk := talk.Talk{
			Handle:  "test",
			Speaker: userID,
		}

		createdConfa, err := confaCRUD.Create(ctx, userID, requestConfa)
		require.NoError(t, err)

		createdTalk, err := talkCRUD.Create(ctx, createdConfa.ID, userID, requestTalk)
		require.NoError(t, err)

		requestClap := Clap{
			Confa: createdConfa.ID,
			Talk:  createdTalk.ID,
			Claps: 2,
		}
		err = clapCRUD.CreateOrUpdate(ctx, userID, requestClap)
		require.NoError(t, err)
		lookup := Lookup{
			Confa: createdConfa.ID,
		}
		claps, err := clapCRUD.Aggregate(ctx, lookup)
		require.NoError(t, err)
		require.Equal(t, claps, 2)
	})
}

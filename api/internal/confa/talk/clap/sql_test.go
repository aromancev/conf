package clap

import (
	"context"
	"github.com/aromancev/confa/internal/confa/talk"

	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/platform/psql"

	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type dependencies struct {
		pg       psql.Conn
		confaSQL *confa.SQL
		talkSQL  *talk.SQL
		clapSQL  *SQL
	}
	setup := func() (dependencies, func()) {
		pg, done := double.NewDocker("", migrate)
		return dependencies{
			pg:       pg,
			confaSQL: confa.NewSQL(),
			talkSQL:  talk.NewSQL(),
			clapSQL:  NewSQL(),
		}, done
	}
	t.Run("Create", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			deps, done := setup()
			defer done()
			ownerId := uuid.New()
			requestConfa := confa.Confa{
				ID:     uuid.New(),
				Owner:  ownerId,
				Handle: "test1",
			}
			_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfa)
			require.NoError(t, err)

			requestTalk := talk.Talk{
				ID:     uuid.New(),
				Owner:  ownerId,
				Confa:  requestConfa.ID,
				Handle: "test1",
			}

			_, err = deps.talkSQL.Create(ctx, deps.pg, requestTalk)
			require.NoError(t, err)

			requestClap := Clap{
				ID:     uuid.New(),
				Owner:  ownerId,
				Speaker:  ownerId,
				Confa:  requestConfa.ID,
				Talk: requestTalk.ID,
				Claps: 5,
			}
			_, err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, requestClap)
			require.NoError(t, err)

		})
		t.Run("Create and Update", func(t *testing.T) {
			t.Parallel()

			deps, done := setup()
			defer done()
			ownerId := uuid.New()
			requestConfa := confa.Confa{
				ID:     uuid.New(),
				Owner:  ownerId,
				Handle: "test1",
			}
			_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfa)
			require.NoError(t, err)

			requestTalk := talk.Talk{
				ID:     uuid.New(),
				Owner:  ownerId,
				Confa:  requestConfa.ID,
				Handle: "test1",
			}

			_, err = deps.talkSQL.Create(ctx, deps.pg, requestTalk)
			require.NoError(t, err)
			clapId := uuid.New()
			requestClap := Clap{
				ID:     clapId,
				Owner:  ownerId,
				Speaker:  ownerId,
				Confa:  requestConfa.ID,
				Talk: requestTalk.ID,
				Claps: 5,
			}
			requestClapUpdate := Clap{
				ID:     clapId,
				Owner:  ownerId,
				Speaker:  ownerId,
				Confa:  requestConfa.ID,
				Talk: requestTalk.ID,
				Claps: 30,
			}
			_, err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, requestClap)
			require.NoError(t, err)
			_, err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, requestClapUpdate)
			require.NoError(t, err)

		})
	})
}

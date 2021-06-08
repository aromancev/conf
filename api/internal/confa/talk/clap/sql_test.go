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
			ownerID := uuid.New()
			requestConfa := confa.Confa{
				ID:     uuid.New(),
				Owner:  ownerID,
				Handle: "test1",
			}
			_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfa)
			require.NoError(t, err)

			requestTalk := talk.Talk{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Handle:  "test1",
			}

			_, err = deps.talkSQL.Create(ctx, deps.pg, requestTalk)
			require.NoError(t, err)

			requestClap := Clap{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Talk:    requestTalk.ID,
				Claps:   5,
			}
			err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, requestClap)
			require.NoError(t, err)

		})
		t.Run("Create and Update", func(t *testing.T) {
			t.Parallel()

			deps, done := setup()
			defer done()
			ownerID := uuid.New()
			requestConfa := confa.Confa{
				ID:     uuid.New(),
				Owner:  ownerID,
				Handle: "test1",
			}
			_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfa)
			require.NoError(t, err)

			requestTalk := talk.Talk{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Handle:  "test1",
			}

			_, err = deps.talkSQL.Create(ctx, deps.pg, requestTalk)
			require.NoError(t, err)
			clapID := uuid.New()
			requestClap := Clap{
				ID:      clapID,
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Talk:    requestTalk.ID,
				Claps:   5,
			}
			requestClapUpdate := Clap{
				ID:      clapID,
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Talk:    requestTalk.ID,
				Claps:   30,
			}
			err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, requestClap)
			require.NoError(t, err)
			err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, requestClapUpdate)
			require.NoError(t, err)

		})
		t.Run("Create and Aggregate", func(t *testing.T) {
			t.Parallel()

			deps, done := setup()
			defer done()
			ownerID := uuid.New()
			requestConfa := confa.Confa{
				ID:     uuid.New(),
				Owner:  ownerID,
				Handle: "test1",
			}
			_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfa)
			require.NoError(t, err)

			requestTalk := talk.Talk{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Handle:  "test1",
			}

			_, err = deps.talkSQL.Create(ctx, deps.pg, requestTalk)
			require.NoError(t, err)
			clapID := uuid.New()
			requestClap := Clap{
				ID:      clapID,
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Talk:    requestTalk.ID,
				Claps:   int8(5),
			}
			clapLookup := Lookup{Talk: requestTalk.ID}
			err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, requestClap)
			require.NoError(t, err)
			claps, err := deps.clapSQL.Aggregate(ctx, deps.pg, clapLookup)
			require.NoError(t, err)
			require.Equal(t, 5, claps)
		})

		t.Run("Create, Update and Aggregate", func(t *testing.T) {
			t.Parallel()

			deps, done := setup()
			defer done()
			ownerID := uuid.New()
			requestConfa := confa.Confa{
				ID:     uuid.New(),
				Owner:  ownerID,
				Handle: "test1",
			}
			_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfa)
			require.NoError(t, err)

			requestTalk := talk.Talk{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Handle:  "test1",
			}

			_, err = deps.talkSQL.Create(ctx, deps.pg, requestTalk)
			require.NoError(t, err)
			clapID := uuid.New()
			requestClap := Clap{
				ID:      clapID,
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Talk:    requestTalk.ID,
				Claps:   int8(5),
			}
			requestClapUpdate := Clap{
				ID:      clapID,
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Talk:    requestTalk.ID,
				Claps:   int8(30),
			}
			clapLookup := Lookup{Talk: requestTalk.ID}
			err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, requestClap)
			require.NoError(t, err)
			claps, err := deps.clapSQL.Aggregate(ctx, deps.pg, clapLookup)
			require.NoError(t, err)
			require.Equal(t, 5, claps)
			err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, requestClapUpdate)
			require.NoError(t, err)
			claps, err = deps.clapSQL.Aggregate(ctx, deps.pg, clapLookup)
			require.NoError(t, err)
			require.Equal(t, 30, claps)

		})
		t.Run("Aggregate by Confa", func(t *testing.T) {
			t.Parallel()

			deps, done := setup()
			defer done()
			ownerID := uuid.New()
			requestConfa := confa.Confa{
				ID:     uuid.New(),
				Owner:  ownerID,
				Handle: "test1",
			}
			_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfa)
			require.NoError(t, err)

			requestTalk := talk.Talk{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Handle:  "test1",
			}
			anotherRequestTalk := talk.Talk{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Handle:  "test2",
			}

			_, err = deps.talkSQL.Create(ctx, deps.pg, requestTalk)
			require.NoError(t, err)
			_, err = deps.talkSQL.Create(ctx, deps.pg, anotherRequestTalk)
			require.NoError(t, err)

			requestClap := Clap{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Talk:    requestTalk.ID,
				Claps:   int8(5),
			}
			anotherRequestClap := Clap{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Talk:    anotherRequestTalk.ID,
				Claps:   int8(30),
			}
			clapLookup := Lookup{Confa: requestConfa.ID}
			err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, requestClap)
			require.NoError(t, err)
			err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, anotherRequestClap)
			require.NoError(t, err)
			claps, err := deps.clapSQL.Aggregate(ctx, deps.pg, clapLookup)
			require.NoError(t, err)
			require.Equal(t, 35, claps)
		})
		t.Run("Aggregate by Speaker", func(t *testing.T) {
			t.Parallel()

			deps, done := setup()
			defer done()
			ownerID := uuid.New()
			requestConfa := confa.Confa{
				ID:     uuid.New(),
				Owner:  ownerID,
				Handle: "test1",
			}
			_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfa)
			require.NoError(t, err)

			requestTalk := talk.Talk{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Handle:  "test1",
			}
			anotherRequestTalk := talk.Talk{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Handle:  "test2",
			}

			_, err = deps.talkSQL.Create(ctx, deps.pg, requestTalk)
			require.NoError(t, err)
			_, err = deps.talkSQL.Create(ctx, deps.pg, anotherRequestTalk)
			require.NoError(t, err)

			requestClap := Clap{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Talk:    requestTalk.ID,
				Claps:   int8(5),
			}
			anotherRequestClap := Clap{
				ID:      uuid.New(),
				Owner:   ownerID,
				Speaker: ownerID,
				Confa:   requestConfa.ID,
				Talk:    anotherRequestTalk.ID,
				Claps:   int8(5),
			}
			clapLookup := Lookup{Speaker: ownerID}
			err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, requestClap)
			require.NoError(t, err)
			err = deps.clapSQL.CreateOrUpdate(ctx, deps.pg, anotherRequestClap)
			require.NoError(t, err)
			claps, err := deps.clapSQL.Aggregate(ctx, deps.pg, clapLookup)
			require.NoError(t, err)
			require.Equal(t, 10, claps)
		})
	})
}

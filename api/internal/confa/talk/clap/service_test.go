package clap

import (
	"context"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/aromancev/confa/internal/platform/psql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestCRUD(t *testing.T) {
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

	t.Run("CreateOrUpdate", func(t *testing.T) {
		t.Parallel()
		deps, done := setup()
		defer done()
		clapCRUD := NewCRUD(deps.pg, NewSQL(), deps.talkSQL)

		userID := uuid.New()

		requestConfa := confa.Confa{
			ID:     uuid.New(),
			Owner:  userID,
			Handle: "test1",
		}
		_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfa)
		require.NoError(t, err)

		requestTalk := talk.Talk{
			ID:      uuid.New(),
			Owner:   userID,
			Speaker: userID,
			Confa:   requestConfa.ID,
			Handle:  "test1",
		}
		_, err = deps.talkSQL.Create(ctx, deps.pg, requestTalk)
		require.NoError(t, err)

		requestClap := Clap{
			Talk:  requestTalk.ID,
			Claps: 2,
		}
		err = clapCRUD.CreateOrUpdate(ctx, userID, requestClap)
		require.NoError(t, err)
		lookup := Lookup{
			Talk: requestTalk.ID,
		}
		claps, err := clapCRUD.Aggregate(ctx, lookup)
		require.NoError(t, err)
		require.Equal(t, 2, claps)
	})

	t.Run("Aggregate", func(t *testing.T) {
		pg, done := double.NewDocker("", migrate)
		defer done()
		talkSQL := talk.NewSQL()
		clapCRUD := NewCRUD(pg, NewSQL(), talkSQL)
		confaCRUD := confa.NewCRUD(pg, confa.NewSQL())
		talkCRUD := talk.NewCRUD(pg, talk.NewSQL(), confaCRUD)

		userID := uuid.New()
		userID2 := uuid.New()
		requestConfa := confa.Confa{
			Handle: "test",
		}
		requestTalk := talk.Talk{
			Handle:  "test",
			Speaker: userID,
		}
		requestTalk2 := talk.Talk{
			Handle:  "test2",
			Speaker: userID,
		}

		createdConfa, err := confaCRUD.Create(ctx, userID, requestConfa)
		require.NoError(t, err)

		createdTalk, err := talkCRUD.Create(ctx, createdConfa.ID, userID, requestTalk)
		require.NoError(t, err)

		createdTalk2, err := talkCRUD.Create(ctx, createdConfa.ID, userID, requestTalk2)
		require.NoError(t, err)

		requestClap := Clap{
			Talk:  createdTalk.ID,
			Claps: 1,
		}
		err = clapCRUD.CreateOrUpdate(ctx, userID, requestClap)
		require.NoError(t, err)

		requestClap2 := Clap{
			Talk:  createdTalk2.ID,
			Claps: 2,
		}
		err = clapCRUD.CreateOrUpdate(ctx, userID2, requestClap2)
		require.NoError(t, err)
		var tests = []struct {
			name string
			in   Lookup
			out  int
		}{
			{"ByTalk", Lookup{Talk: createdTalk.ID}, 1},
			{"ByConfa", Lookup{Confa: createdConfa.ID}, 3},
			{"BySpeaker", Lookup{Speaker: createdTalk.Speaker}, 3},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				claps, err := clapCRUD.Aggregate(ctx, tt.in)
				require.NoError(t, err)
				require.Equal(t, tt.out, claps)
			})
		}
	})

}

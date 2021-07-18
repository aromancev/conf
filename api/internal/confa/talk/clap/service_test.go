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
		_, err = clapCRUD.CreateOrUpdate(ctx, userID, requestClap)
		require.NoError(t, err)
		lookup := Lookup{
			Talk: requestTalk.ID,
		}
		claps, err := clapCRUD.Aggregate(ctx, lookup)
		require.NoError(t, err)
		require.Equal(t, 2, claps)
	})

	t.Run("Aggregate", func(t *testing.T) {
		t.Parallel()
		deps, done := setup()
		defer done()
		clapCRUD := NewCRUD(deps.pg, NewSQL(), deps.talkSQL)

		var ids [2]uuid.UUID
		for i := range ids {
			ids[i] = uuid.New()
		}

		requestConfas := []confa.Confa{
			{
				ID:     uuid.New(),
				Owner:  ids[0],
				Handle: "test1",
			},
			{
				ID:     uuid.New(),
				Owner:  ids[0],
				Handle: "test2",
			},
			{
				ID:     uuid.New(),
				Owner:  ids[1],
				Handle: "test3",
			},
		}

		_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfas...)
		require.NoError(t, err)
		requestTalks := []talk.Talk{
			{
				ID:      uuid.New(),
				Owner:   ids[0],
				Speaker: ids[0],
				Confa:   requestConfas[0].ID,
				Handle:  "test1",
			},
			{
				ID:      uuid.New(),
				Owner:   ids[0],
				Speaker: ids[0],
				Confa:   requestConfas[1].ID,
				Handle:  "test2",
			},
			{
				ID:      uuid.New(),
				Owner:   ids[1],
				Speaker: ids[1],
				Confa:   requestConfas[2].ID,
				Handle:  "test3",
			},
		}

		_, err = deps.talkSQL.Create(ctx, deps.pg, requestTalks...)
		require.NoError(t, err)

		_, err = clapCRUD.CreateOrUpdate(ctx, ids[0], Clap{Talk: requestTalks[0].ID, Claps: 1})
		require.NoError(t, err)
		_, err = clapCRUD.CreateOrUpdate(ctx, ids[1], Clap{Talk: requestTalks[0].ID, Claps: 1})
		require.NoError(t, err)
		_, err = clapCRUD.CreateOrUpdate(ctx, ids[0], Clap{Talk: requestTalks[1].ID, Claps: 1})
		require.NoError(t, err)
		_, err = clapCRUD.CreateOrUpdate(ctx, ids[1], Clap{Talk: requestTalks[2].ID, Claps: 1})
		require.NoError(t, err)

		t.Run("bySpeaker", func(t *testing.T) {
			lookup := Lookup{
				Speaker: ids[0],
			}
			claps, err := clapCRUD.Aggregate(ctx, lookup)
			require.NoError(t, err)
			require.Equal(t, 3, claps)
			lookup = Lookup{
				Speaker: ids[1],
			}
			claps, err = clapCRUD.Aggregate(ctx, lookup)
			require.NoError(t, err)
			require.Equal(t, 1, claps)
		})

		t.Run("byTalk", func(t *testing.T) {
			lookup := Lookup{
				Talk: requestTalks[0].ID,
			}
			claps, err := clapCRUD.Aggregate(ctx, lookup)
			require.NoError(t, err)
			require.Equal(t, 2, claps)
			lookup = Lookup{
				Talk: requestTalks[1].ID,
			}
			claps, err = clapCRUD.Aggregate(ctx, lookup)
			require.NoError(t, err)
			require.Equal(t, 1, claps)
			lookup = Lookup{
				Talk: requestTalks[2].ID,
			}
			claps, err = clapCRUD.Aggregate(ctx, lookup)
			require.NoError(t, err)
			require.Equal(t, 1, claps)
		})

		t.Run("byConfa", func(t *testing.T) {
			lookup := Lookup{
				Confa: requestConfas[0].ID,
			}
			claps, err := clapCRUD.Aggregate(ctx, lookup)
			require.NoError(t, err)
			require.Equal(t, 2, claps)
			lookup = Lookup{
				Confa: requestConfas[1].ID,
			}
			claps, err = clapCRUD.Aggregate(ctx, lookup)
			require.NoError(t, err)
			require.Equal(t, 1, claps)
			lookup = Lookup{
				Confa: requestConfas[2].ID,
			}
			claps, err = clapCRUD.Aggregate(ctx, lookup)
			require.NoError(t, err)
			require.Equal(t, 1, claps)
		})
	})
}

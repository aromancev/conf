package talk

import (
	"context"

	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/platform/psql"

	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type dependencies struct {
		pg       psql.Conn
		confaSQL *confa.SQL
		talkSQL  *SQL
	}
	setup := func() (dependencies, func()) {
		pg, done := double.NewDocker("", migrate)
		return dependencies{
			pg:       pg,
			confaSQL: confa.NewSQL(),
			talkSQL:  NewSQL(),
		}, done
	}
	t.Run("Create", func(t *testing.T) {
		t.Run("String handle", func(t *testing.T) {
			t.Parallel()

			deps, done := setup()
			defer done()
			userID := uuid.New()

			requestConfa := confa.Confa{
				ID:     uuid.New(),
				Owner:  userID,
				Handle: "test1",
			}
			_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfa)
			require.NoError(t, err)

			requestTalk := Talk{
				ID:     uuid.New(),
				Owner:  userID,
				Speaker:  userID,
				Confa:  requestConfa.ID,
				Handle: "test1",
			}

			createdTalk, err := deps.talkSQL.Create(ctx, deps.pg, requestTalk)
			require.NoError(t, err)

			fetchedTalk, err := deps.talkSQL.Fetch(ctx, deps.pg, Lookup{
				ID:    requestTalk.ID,
				Confa: requestTalk.Confa,
			})
			require.NoError(t, err)
			assert.Equal(t, createdTalk, fetchedTalk)
		})
		t.Run("UUID handle", func(t *testing.T) {
			t.Parallel()

			deps, done := setup()
			defer done()
			userID := uuid.New()

			requestConfa := confa.Confa{
				ID:     uuid.New(),
				Owner:  userID,
				Handle: "test2",
			}
			_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfa)
			require.NoError(t, err)

			requestTalk := Talk{
				ID:     uuid.New(),
				Owner:  userID,
				Speaker:  userID,
				Confa:  requestConfa.ID,
				Handle: uuid.New().String(),
			}

			createdTalk, err := deps.talkSQL.Create(ctx, deps.pg, requestTalk)
			require.NoError(t, err)

			fetchedTalk, err := deps.talkSQL.Fetch(ctx, deps.pg, Lookup{
				ID:    requestTalk.ID,
				Confa: requestTalk.Confa,
			})
			require.NoError(t, err)
			assert.Equal(t, createdTalk, fetchedTalk)
		})
		t.Run("Duplicated Entry", func(t *testing.T) {
			t.Parallel()

			deps, done := setup()
			defer done()
			userID := uuid.New()

			requestConfa := confa.Confa{
				ID:     uuid.New(),
				Owner:  userID,
				Handle: "test3",
			}
			_, err := deps.confaSQL.Create(ctx, deps.pg, requestConfa)
			require.NoError(t, err)

			requestTalk := Talk{
				ID:     uuid.New(),
				Owner:  userID,
				Speaker:  userID,
				Confa:  requestConfa.ID,
				Handle: "test3",
			}

			_, err = deps.talkSQL.Create(ctx, deps.pg, requestTalk)
			require.NoError(t, err)
			_, err = deps.talkSQL.Create(ctx, deps.pg, requestTalk)
			assert.ErrorIs(t, err, ErrDuplicatedEntry)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()
		talks := NewSQL()
		confas := confa.NewSQL()
		userID := uuid.New()

		conf := confa.Confa{
			ID:     uuid.New(),
			Owner:  userID,
			Handle: "test",
		}

		tlk := Talk{
			ID:     uuid.New(),
			Owner:  userID,
			Speaker:  userID,
			Confa:  conf.ID,
			Handle: "test",
		}

		_, err := confas.Create(ctx, pg, conf)
		require.NoError(t, err)

		createdTalk, err := talks.Create(ctx, pg, tlk)
		require.NoError(t, err)

		t.Run("ID", func(t *testing.T) {
			fetchedTalk, err := talks.Fetch(ctx, pg, Lookup{
				ID: tlk.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, createdTalk, fetchedTalk)
		})

		t.Run("Confa", func(t *testing.T) {
			fetchedTalk, err := talks.Fetch(ctx, pg, Lookup{
				Confa: conf.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, createdTalk, fetchedTalk)
		})
	})
}

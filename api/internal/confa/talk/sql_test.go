package talk

import (
	"context"
	"github.com/aromancev/confa/internal/confa"

	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()
		confas := confa.NewSQL()

		requestConfa := confa.Confa{
			ID:     uuid.New(),
			Owner:  uuid.New(),
			Handle: "test",
		}
		_, err := confas.Create(ctx, pg, requestConfa)
		require.NoError(t, err)

		talks := NewSQL()

		requestTalk := Talk{
			ID:     uuid.New(),
			Confa:  requestConfa.ID,
			Handle: "test",
		}

		createdTalk, err := talks.Create(ctx, pg, requestTalk)
		require.NoError(t, err)

		fetchedTalk, err := talks.Fetch(ctx, pg, Lookup{
			ID:    requestTalk.ID,
			Confa: requestTalk.Confa,
		})
		assert.Equal(t, createdTalk, fetchedTalk)
	})

	t.Run("Handle-UUID", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()
		confas := confa.NewSQL()

		requestConfa := confa.Confa{
			ID:     uuid.New(),
			Owner:  uuid.New(),
			Handle: uuid.New().String(),
		}
		_, err := confas.Create(ctx, pg, requestConfa)
		require.NoError(t, err)

		talks := NewSQL()

		requestTalk := Talk{
			ID:     uuid.New(),
			Confa:  requestConfa.ID,
			Handle: uuid.New().String(),
		}

		createdTalk, err := talks.Create(ctx, pg, requestTalk)
		require.NoError(t, err)

		fetchedTalk, err := talks.Fetch(ctx, pg, Lookup{
			ID:    requestTalk.ID,
			Confa: requestTalk.Confa,
		})
		assert.Equal(t, createdTalk, fetchedTalk)
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()
		talks := NewSQL()
		confas := confa.NewSQL()

		conf := confa.Confa{
			ID:     uuid.New(),
			Owner:  uuid.New(),
			Handle: "test",
		}

		tlk := Talk{
			ID:     uuid.New(),
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

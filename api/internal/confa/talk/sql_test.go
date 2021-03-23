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
		sqlConfa := confa.NewSQL()

		requestConfa := confa.Confa{
			ID:     uuid.New(),
			Owner:  uuid.New(),
			Handle: "test",
		}
		_, err := sqlConfa.Create(ctx, pg, requestConfa)
		require.NoError(t, err)

		sqlTalk := NewSQL()

		requestTalk := Talk{
			ID:     uuid.New(),
			Confa:  requestConfa.ID,
			Handle: "test",
		}

		createdTalk, err := sqlTalk.Create(ctx, pg, requestTalk)
		require.NoError(t, err)

		fetchedTalk, err := sqlTalk.Fetch(ctx, pg, Lookup{
			ID:    requestTalk.ID,
			Confa: requestTalk.Confa,
		})
		assert.Equal(t, createdTalk, fetchedTalk)
	})

	t.Run("Handle-UUID", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()
		sqlConfa := confa.NewSQL()

		requestConfa := confa.Confa{
			ID:     uuid.New(),
			Owner:  uuid.New(),
			Handle: uuid.New().String(),
		}
		_, err := sqlConfa.Create(ctx, pg, requestConfa)
		require.NoError(t, err)

		sqlTalk := NewSQL()

		requestTalk := Talk{
			ID:     uuid.New(),
			Confa:  requestConfa.ID,
			Handle: uuid.New().String(),
		}

		createdTalk, err := sqlTalk.Create(ctx, pg, requestTalk)
		require.NoError(t, err)

		fetchedTalk, err := sqlTalk.Fetch(ctx, pg, Lookup{
			ID:    requestTalk.ID,
			Confa: requestTalk.Confa,
		})
		assert.Equal(t, createdTalk, fetchedTalk)
	})
}

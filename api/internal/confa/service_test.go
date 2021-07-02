package confa

import (
	"context"
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

		crud := NewCRUD(pg, NewSQL())
		confa, err := crud.Create(ctx, uuid.New(), Confa{
			Handle: "test",
		})
		require.NoError(t, err)
		fetched, err := crud.FetchOne(ctx, Lookup{ID: confa.ID})
		require.NoError(t, err)
		require.Equal(t, confa, fetched)
	})
}

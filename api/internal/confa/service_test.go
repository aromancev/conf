package confa

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("SetGet", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker(migrations)
		defer done()
		crud := NewCRUD(pg, NewSQL())
		userid := uuid.New()
		request := Confa{
			Handle: "test",
		}

		confa, err := crud.Create(ctx, userid, request)
		require.NoError(t, err)
		fetchedConfa, err := crud.Fetch(ctx, confa.ID, confa.Owner)
		require.NoError(t, err)
		require.Equal(t, confa, fetchedConfa)
	})

}

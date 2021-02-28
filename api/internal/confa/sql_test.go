package confa

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

const (
	migrations = "../migrations"
)

func TestSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker(migrations)
		defer done()
		sql := NewSQL()

		request := Confa{
			ID:    uuid.New(),
			Owner: uuid.New(),
		}
		created, err := sql.Create(ctx, pg, request)
		require.NoError(t, err)
		assert.Equal(t, request, created)
	})
}

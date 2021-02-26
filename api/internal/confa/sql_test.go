package confa

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestSQL(t *testing.T) {
	t.Parallel()

	db, done := double.NewDocker(t, "")
	defer done()

	_, err := db.QueryContext(context.Background(), "select 1")
	require.NoError(t, err)
}

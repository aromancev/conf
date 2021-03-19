package ident

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
	"github.com/aromancev/confa/internal/user"
)

func TestCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("GetOrCreate", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()

		idents := NewSQL()
		crud := NewCRUD(pg, idents, user.NewSQL())

		with := Ident{
			Platform: PlatformEmail,
			Value:    "test",
		}

		first, err := crud.GetOrCreate(ctx, with)
		require.NoError(t, err)

		second, err := crud.GetOrCreate(ctx, with)
		require.NoError(t, err)

		assert.Equal(t, first, second)

		ident, err := idents.FetchOne(ctx, pg, Lookup{})
		require.NoError(t, err)
		assert.Equal(t, first, ident.Owner)
		assert.Equal(t, with.Platform, ident.Platform)
		assert.Equal(t, with.Value, ident.Value)
	})
}

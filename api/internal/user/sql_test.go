package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TesSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()

		users := NewSQL()

		user := User{ID: uuid.New()}
		createdUser, _ := users.Create(ctx, pg, user)

		t.Run("Fetch", func(t *testing.T) {
			fetchedUser, err := users.Fetch(ctx, pg, Lookup{
				ID: user.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, createdUser, fetchedUser)
		})

		t.Run("FetchOne", func(t *testing.T) {
			fetchedUser, err := users.FetchOne(ctx, pg, Lookup{
				ID: user.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, createdUser[0], fetchedUser)
		})
	})
}

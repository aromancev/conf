package clap

import (
	"context"
	"testing"

	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("CreateOrUpdate", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			db := dockerMongo(t)
			talkMongo := talk.NewMongo(db)
			clapMongo := NewMongo(db)
			clapCRUD := NewCRUD(clapMongo, talkMongo)

			tlk := talk.Talk{
				ID:      uuid.New(),
				Confa:   uuid.New(),
				Owner:   uuid.New(),
				Speaker: uuid.New(),
				Room:    uuid.New(),
				Handle:  "test",
			}

			_, err := talkMongo.Create(ctx, tlk)
			require.NoError(t, err)

			id, err := clapCRUD.CreateOrUpdate(ctx, uuid.New(), tlk.ID, 1)
			require.NoError(t, err)
			assert.NotZero(t, id)
		})

		t.Run("Non existent talk returns error", func(t *testing.T) {
			t.Parallel()

			db := dockerMongo(t)
			talkMongo := talk.NewMongo(db)
			clapMongo := NewMongo(db)
			clapCRUD := NewCRUD(clapMongo, talkMongo)

			_, err := clapCRUD.CreateOrUpdate(ctx, uuid.New(), uuid.New(), 1)
			require.ErrorIs(t, err, talk.ErrNotFound)
		})
	})
}

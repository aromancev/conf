package talk

import (
	"context"
	"testing"

	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/proto/rtc/double"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			db := dockerMongo(t)
			confaMongo := confa.NewMongo(db)
			rooms := double.NewMemory()
			talkCRUD := NewCRUD(NewMongo(db), confaMongo, rooms)

			conf := confa.Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test",
			}
			_, err := confaMongo.Create(ctx, conf)
			require.NoError(t, err)

			created, err := talkCRUD.Create(ctx, conf.Owner, Talk{
				Confa:  conf.ID,
				Handle: "test",
			})
			require.NoError(t, err)
			fetched, err := talkCRUD.Fetch(ctx, Lookup{ID: created.ID})
			require.NoError(t, err)
			assert.Equal(t, []Talk{created}, fetched)

			room, err := rooms.Room(ctx, fetched[0].Room.String())
			require.NoError(t, err)
			assert.Equal(t, fetched[0].Room.String(), room.Id)
		})

		t.Run("Only the owner can create", func(t *testing.T) {
			db := dockerMongo(t)
			confaMongo := confa.NewMongo(db)
			talkCRUD := NewCRUD(NewMongo(db), confaMongo, nil)

			conf := confa.Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test",
			}
			_, err := confaMongo.Create(ctx, conf)
			require.NoError(t, err)

			_, err = talkCRUD.Create(ctx, uuid.New(), Talk{
				Confa:  conf.ID,
				Handle: "test",
			})
			require.ErrorIs(t, err, ErrPermissionDenied)
		})
	})
}

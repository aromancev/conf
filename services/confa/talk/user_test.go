package talk

import (
	"context"
	"testing"
	"time"

	"github.com/aromancev/confa/confa"
	"github.com/aromancev/confa/internal/proto/rtc/double"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type emitterStub struct {
}

func (s *emitterStub) StartRecording(ctx context.Context, talkID, roomID uuid.UUID) error {
	return nil
}

func (s *emitterStub) StopRecording(ctx context.Context, talkID, roomID uuid.UUID, after time.Duration) error {
	return nil
}

func TestUserService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			db := dockerMongo(t)
			confaMongo := confa.NewMongo(db)
			rooms := double.NewMemory()
			service := NewUserService(NewMongo(db), confaMongo, &emitterStub{}, rooms)

			conf := confa.Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test",
			}
			_, err := confaMongo.Create(ctx, conf)
			require.NoError(t, err)

			created, err := service.Create(
				ctx, conf.Owner,
				confa.Lookup{
					ID: conf.ID,
				},
				Talk{
					Handle: "test",
				},
			)
			require.NoError(t, err)
			fetched, err := service.Fetch(ctx, Lookup{ID: created.ID})
			require.NoError(t, err)
			assert.Equal(t, []Talk{created}, fetched)

			room, err := rooms.Room(ctx, fetched[0].Room.String())
			require.NoError(t, err)
			assert.Equal(t, fetched[0].Room[:], room.Id)
		})

		t.Run("Only the owner can create", func(t *testing.T) {
			db := dockerMongo(t)
			confaMongo := confa.NewMongo(db)
			service := NewUserService(NewMongo(db), confaMongo, &emitterStub{}, nil)

			conf := confa.Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test",
			}
			_, err := confaMongo.Create(ctx, conf)
			require.NoError(t, err)

			_, err = service.Create(
				ctx,
				uuid.New(),
				confa.Lookup{
					ID: conf.ID,
				},
				Talk{
					Handle: "test",
				})
			require.ErrorIs(t, err, confa.ErrNotFound)
		})
	})
}

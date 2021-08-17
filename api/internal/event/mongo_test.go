package event

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventMongo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			events := NewMongo(dockerMongo(t))

			request := Event{
				ID:    uuid.New(),
				Owner: uuid.New(),
				Room:  uuid.New(),
				Payload: Payload{
					Type: TypePeerStatus,
					Payload: PayloadPeerStatus{
						Status: PeerJoined,
					},
				},
			}
			created, err := events.Create(ctx, request)
			require.NoError(t, err)
			assert.NotZero(t, created[0].CreatedAt)

			fetched, err := events.Fetch(ctx, Lookup{
				ID: request.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
		t.Run("Duplicated event returns correct error", func(t *testing.T) {
			t.Parallel()

			events := NewMongo(dockerMongo(t))

			request := Event{
				ID:    uuid.New(),
				Owner: uuid.New(),
				Room:  uuid.New(),
				Payload: Payload{
					Type: TypePeerStatus,
					Payload: PayloadPeerStatus{
						Status: PeerJoined,
					},
				},
			}
			_, err := events.Create(ctx, request)
			require.NoError(t, err)
			_, err = events.Create(ctx, request)
			require.ErrorIs(t, err, ErrDuplicatedEntry)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		events := NewMongo(dockerMongo(t))

		event := Event{
			ID:    uuid.New(),
			Owner: uuid.New(),
			Room:  uuid.New(),
			Payload: Payload{
				Type: TypePeerStatus,
				Payload: PayloadPeerStatus{
					Status: PeerJoined,
				},
			},
		}
		created, err := events.Create(ctx, event)
		require.NoError(t, err)
		_, err = events.Create(
			ctx,
			Event{
				ID:    uuid.New(),
				Owner: uuid.New(),
				Room:  uuid.New(),
				Payload: Payload{
					Type: TypePeerStatus,
					Payload: PayloadPeerStatus{
						Status: PeerJoined,
					},
				},
			},
			Event{
				ID:    uuid.New(),
				Owner: uuid.New(),
				Room:  uuid.New(),
				Payload: Payload{
					Type: TypePeerStatus,
					Payload: PayloadPeerStatus{
						Status: PeerJoined,
					},
				},
			},
		)
		require.NoError(t, err)

		t.Run("by id", func(t *testing.T) {
			fetched, err := events.Fetch(ctx, Lookup{
				ID: event.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("by room", func(t *testing.T) {
			fetched, err := events.Fetch(ctx, Lookup{
				Room: event.Room,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("with limit and offset", func(t *testing.T) {
			fetched, err := events.Fetch(ctx, Lookup{
				Limit: 1,
			})
			require.NoError(t, err)
			assert.Equal(t, 1, len(fetched))

			// 3 in total, skipped one.
			fetched, err = events.Fetch(ctx, Lookup{
				From: fetched[0].ID,
			})
			require.NoError(t, err)
			assert.Equal(t, 2, len(fetched))
		})
	})

	t.Run("Watch", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			events := NewMongo(dockerMongo(t))

			request := Event{
				ID:    uuid.New(),
				Owner: uuid.New(),
				Room:  uuid.New(),
				Payload: Payload{
					Type: TypePeerStatus,
					Payload: PayloadPeerStatus{
						Status: PeerJoined,
					},
				},
			}
			cur, err := events.Watch(ctx, request.Room)
			require.NoError(t, err)

			created, err := events.CreateOne(ctx, request)
			require.NoError(t, err)

			ev, err := cur.Next(ctx)
			require.NoError(t, err)
			assert.Equal(t, created, ev)
		})
		t.Run("Filters by room", func(t *testing.T) {
			t.Parallel()

			events := NewMongo(dockerMongo(t))

			request := Event{
				ID:    uuid.New(),
				Owner: uuid.New(),
				Room:  uuid.New(),
				Payload: Payload{
					Type: TypePeerStatus,
					Payload: PayloadPeerStatus{
						Status: PeerJoined,
					},
				},
			}
			cur, err := events.Watch(ctx, uuid.New())
			require.NoError(t, err)

			_, err = events.CreateOne(ctx, request)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err = cur.Next(ctx)
			require.Error(t, err)
		})
	})
}

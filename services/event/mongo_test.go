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
				ID:   uuid.New(),
				Room: uuid.New(),
				Payload: Payload{
					PeerState: &PayloadPeerState{
						Peer:      uuid.New(),
						SessionID: uuid.New(),
						Status:    PeerJoined,
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
				ID:   uuid.New(),
				Room: uuid.New(),
				Payload: Payload{
					PeerState: &PayloadPeerState{
						Peer:      uuid.New(),
						SessionID: uuid.New(),
						Status:    PeerJoined,
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
		roomID := uuid.New()

		created, err := events.Create(
			ctx,
			Event{
				ID:   uuid.UUID{1},
				Room: roomID,
				Payload: Payload{
					PeerState: &PayloadPeerState{
						Peer:      uuid.New(),
						SessionID: uuid.New(),
						Status:    PeerJoined,
					},
				},
			},
			Event{
				ID:   uuid.UUID{2},
				Room: roomID,
				Payload: Payload{
					PeerState: &PayloadPeerState{
						Peer:      uuid.New(),
						SessionID: uuid.New(),
						Status:    PeerJoined,
					},
				},
			},
			Event{
				ID:   uuid.UUID{3},
				Room: roomID,
				Payload: Payload{
					PeerState: &PayloadPeerState{
						Peer:      uuid.New(),
						SessionID: uuid.New(),
						Status:    PeerJoined,
					},
				},
			},
		)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)
		createdLater, err := events.Create(
			ctx,
			Event{
				ID:   uuid.UUID{4},
				Room: roomID,
				Payload: Payload{
					PeerState: &PayloadPeerState{
						Peer:      uuid.New(),
						SessionID: uuid.New(),
						Status:    PeerJoined,
					},
				},
			},
		)
		require.NoError(t, err)

		t.Run("By id", func(t *testing.T) {
			fetched, err := events.Fetch(ctx, Lookup{
				ID: created[0].ID,
			})
			require.NoError(t, err)
			assert.Equal(t, []Event{created[0]}, fetched)
		})

		t.Run("By room", func(t *testing.T) {
			fetched, err := events.Fetch(ctx, Lookup{
				Room: roomID,
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, append(created, createdLater...), fetched)
		})

		t.Run("Pagination works in both directions", func(t *testing.T) {
			fetched, err := events.Fetch(ctx, Lookup{
				From: From{
					ID: created[1].ID,
				},
				Limit: 1,
				Asc:   true,
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, []Event{created[2]}, fetched)

			fetched, err = events.Fetch(ctx, Lookup{
				From: From{
					ID: created[1].ID,
				},
				Limit: 1,
				Asc:   false,
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, []Event{created[0]}, fetched)
		})

		t.Run("Pagination with createdAt works", func(t *testing.T) {
			fetched, err := events.Fetch(ctx, Lookup{
				From: From{
					ID:        created[2].ID,
					CreatedAt: created[2].CreatedAt,
				},
				Limit: 1,
				Asc:   true,
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, createdLater, fetched)
		})
	})

	t.Run("Watch", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			events := NewMongo(dockerMongo(t))

			request := Event{
				ID:   uuid.New(),
				Room: uuid.New(),
				Payload: Payload{
					PeerState: &PayloadPeerState{
						Peer:      uuid.New(),
						SessionID: uuid.New(),
						Status:    PeerJoined,
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
				ID:   uuid.New(),
				Room: uuid.New(),
				Payload: Payload{
					PeerState: &PayloadPeerState{
						Peer:      uuid.New(),
						SessionID: uuid.New(),
						Status:    PeerJoined,
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

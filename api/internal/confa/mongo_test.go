package confa

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestWatch(t *testing.T) {
	ctx := context.Background()

	db := dockerMongo(t)

	coll := db.Collection("test")

	stream, err := coll.Watch(ctx, mongo.Pipeline{})
	require.NoError(t, err)

	go func() {
		for stream.Next(ctx) {
			fmt.Println(stream.Current)
		}
	}()

	_, err = coll.InsertOne(ctx, bson.M{"a": "b"})
	require.NoError(t, err)

	time.Sleep(time.Second)
}

func TestMongo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		t.Run("String handle", func(t *testing.T) {
			t.Parallel()

			confas := NewMongo(dockerMongo(t))

			request := Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test1",
			}
			created, err := confas.Create(ctx, request)
			require.NoError(t, err)

			fetched, err := confas.Fetch(ctx, Lookup{
				ID:    request.ID,
				Owner: request.Owner,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})
		t.Run("UUID handle", func(t *testing.T) {
			t.Parallel()

			confas := NewMongo(dockerMongo(t))

			request := Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: uuid.New().String(),
			}
			_, err := confas.Create(ctx, request)
			require.NoError(t, err)
		})
		t.Run("Duplicated Entry", func(t *testing.T) {
			t.Parallel()

			confas := NewMongo(dockerMongo(t))

			_, err := confas.Create(ctx, Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test",
			})
			require.NoError(t, err)
			_, err = confas.Create(ctx, Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: "test",
			})
			require.ErrorIs(t, err, ErrDuplicateEntry)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		confas := NewMongo(dockerMongo(t))

		conf := Confa{
			ID:     uuid.New(),
			Owner:  uuid.New(),
			Handle: "test",
		}
		created, err := confas.Create(ctx, conf)
		require.NoError(t, err)
		_, err = confas.Create(
			ctx,
			Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: uuid.NewString(),
			},
			Confa{
				ID:     uuid.New(),
				Owner:  uuid.New(),
				Handle: uuid.NewString(),
			},
		)
		require.NoError(t, err)

		t.Run("by id", func(t *testing.T) {
			fetched, err := confas.Fetch(ctx, Lookup{
				ID: conf.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("by owner", func(t *testing.T) {
			fetched, err := confas.Fetch(ctx, Lookup{
				Owner: conf.Owner,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("with limit and offset", func(t *testing.T) {
			fetched, err := confas.Fetch(ctx, Lookup{
				Limit: 1,
			})
			require.NoError(t, err)
			assert.Equal(t, 1, len(fetched))

			// 3 in total, skipped one.
			fetched, err = confas.Fetch(ctx, Lookup{
				From: fetched[0].ID,
			})
			require.NoError(t, err)
			assert.Equal(t, 2, len(fetched))
		})
	})
}

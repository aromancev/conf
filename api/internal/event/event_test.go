package event

import (
	"encoding/json"
	"testing"

	"github.com/aromancev/confa/internal/platform/mongo/double"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func dockerMongo(t *testing.T) *mongo.Database {
	t.Helper()

	db := double.NewDocker()
	driver, err := mongodb.WithInstance(db.Client(), &mongodb.Config{
		DatabaseName: db.Name(),
	})
	require.NoError(t, err)
	migrator, err := migrate.NewWithDatabaseInstance("file://../migrations/rtc", db.Name(), driver)
	require.NoError(t, err)
	require.NoError(t, migrator.Up())
	return db
}

func TestEvent(t *testing.T) {
	t.Parallel()

	t.Run("JSON Marshall", func(t *testing.T) {
		expected := Event{
			ID:    uuid.New(),
			Owner: uuid.New(),
			Room:  uuid.New(),
			Payload: Payload{
				Type: TypePeerState,
				Payload: PayloadPeerState{
					Status: PeerJoined,
				},
			},
		}
		buf, err := json.Marshal(expected)
		require.NoError(t, err)
		var actual Event
		require.NoError(t, json.Unmarshal(buf, &actual))
		assert.Equal(t, expected, actual)
	})
	t.Run("BSON Marshall", func(t *testing.T) {
		expected := Event{
			ID:    uuid.New(),
			Owner: uuid.New(),
			Room:  uuid.New(),
			Payload: Payload{
				Type: TypePeerState,
				Payload: PayloadPeerState{
					Status: PeerJoined,
				},
			},
		}
		buf, err := bson.Marshal(expected)
		require.NoError(t, err)
		var actual Event
		require.NoError(t, bson.Unmarshal(buf, &actual))
		assert.Equal(t, expected, actual)
	})
}

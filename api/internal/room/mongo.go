package room

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	batchLimit = 100
)

type Mongo struct {
	db *mongo.Database
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		db: db,
	}
}

func (m *Mongo) Create(ctx context.Context, requests ...Room) ([]Room, error) {
	if len(requests) == 0 {
		return nil, errors.New("trying to create zero objects")
	}
	if len(requests) > batchLimit {
		return nil, fmt.Errorf("trying to create more than %d", batchLimit)
	}

	now := mongoNow()
	docs := make([]interface{}, len(requests))
	for i, r := range requests {
		if err := r.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrValidation, err)
		}
		requests[i].CreatedAt = now
		docs[i] = requests[i]
	}

	_, err := m.db.Collection("rooms").InsertMany(ctx, docs)
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (m *Mongo) Fetch(ctx context.Context, lookup Lookup) ([]Room, error) {
	filter := bson.M{}
	switch {
	case lookup.ID != uuid.Nil:
		filter["_id"] = lookup.ID
	case lookup.From != uuid.Nil:
		filter["_id"] = bson.M{
			"$gt": lookup.From,
		}
	}
	if lookup.Owner != uuid.Nil {
		filter["owner"] = lookup.Owner
	}
	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}

	cur, err := m.db.Collection("rooms").Find(
		ctx,
		filter,
		&options.FindOptions{
			Sort:  bson.M{"_id": 1},
			Limit: &lookup.Limit,
		},
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var rooms []Room
	for cur.Next(ctx) {
		var r Room
		err := cur.Decode(&r)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}

	return rooms, nil
}

func (m *Mongo) FetchOne(ctx context.Context, lookup Lookup) (Room, error) {
	confas, err := m.Fetch(ctx, lookup)
	if err != nil {
		return Room{}, err
	}
	if len(confas) == 0 {
		return Room{}, ErrNotFound
	}
	if len(confas) > 1 {
		return Room{}, ErrUnexpectedResult
	}
	return confas[0], nil
}

func mongoNow() time.Time {
	// Mongodb only stores milliseconds.
	return time.Now().UTC().Round(time.Millisecond)
}

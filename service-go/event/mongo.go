package event

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
	batchLimit = 5000
)

type MongoCursor struct {
	stream *mongo.ChangeStream
}

func (c *MongoCursor) Next(ctx context.Context) (Event, error) {
	hasNext := c.stream.Next(ctx)
	if !hasNext {
		if c.stream.Err() != nil {
			return Event{}, c.stream.Err()
		}
		return Event{}, ErrCursorClosed
	}
	var change struct {
		Doc Event `bson:"fullDocument"`
	}
	err := c.stream.Decode(&change)
	if err != nil {
		return Event{}, err
	}
	return change.Doc, nil
}

func (c *MongoCursor) Close(ctx context.Context) error {
	return c.stream.Close(ctx)
}

type Mongo struct {
	db *mongo.Database
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		db: db,
	}
}

func (m *Mongo) Create(ctx context.Context, requests ...Event) ([]Event, error) {
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

	_, err := m.db.Collection("events").InsertMany(ctx, docs)
	switch {
	case mongo.IsDuplicateKeyError(err):
		return nil, ErrDuplicatedEntry
	case err != nil:
		return nil, err
	}
	return requests, nil
}

func (m *Mongo) CreateOne(ctx context.Context, request Event) (Event, error) {
	events, err := m.Create(ctx, request)
	if err != nil {
		return Event{}, err
	}
	if len(events) == 0 {
		return Event{}, ErrNotFound
	}
	if len(events) > 1 {
		return Event{}, ErrUnexpectedResult
	}
	return events[0], nil
}

func (m *Mongo) Fetch(ctx context.Context, lookup Lookup) ([]Event, error) {
	filter := bson.M{}
	switch {
	case lookup.ID != uuid.Nil:
		filter["_id"] = lookup.ID
	case lookup.From.ID != uuid.Nil:
		filter["_id"] = bson.M{
			"$gt": lookup.From.ID,
		}
	case lookup.From.CreatedAt != time.Time{}:
		filter["createdAt"] = bson.M{
			"$gt": lookup.From.CreatedAt,
		}
	}
	if lookup.Room != uuid.Nil {
		filter["roomId"] = lookup.Room
	}
	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}
	order := -1
	if lookup.Asc {
		order = 1
	}

	cur, err := m.db.Collection("events").Find(
		ctx,
		filter,
		&options.FindOptions{
			Sort:  bson.D{{Key: "createdAt", Value: order}, {Key: "_id", Value: order}},
			Limit: &lookup.Limit,
		},
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var events []Event
	for cur.Next(ctx) {
		var e Event
		err := cur.Decode(&e)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}

func (m *Mongo) FetchOne(ctx context.Context, lookup Lookup) (Event, error) {
	confas, err := m.Fetch(ctx, lookup)
	if err != nil {
		return Event{}, err
	}
	if len(confas) == 0 {
		return Event{}, ErrNotFound
	}
	if len(confas) > 1 {
		return Event{}, ErrUnexpectedResult
	}
	return confas[0], nil
}

func (m *Mongo) Watch(ctx context.Context, roomID uuid.UUID) (Cursor, error) {
	stream, err := m.db.Collection("events").Watch(ctx, mongo.Pipeline{
		{{
			Key: "$match",
			Value: bson.M{
				"operationType":       "insert",
				"fullDocument.roomId": roomID,
			},
		}},
		{{
			Key: "$project",
			Value: bson.M{
				"fullDocument": 1,
			},
		}},
	})
	if err != nil {
		return nil, err
	}
	return &MongoCursor{stream: stream}, nil
}

func mongoNow() time.Time {
	// Mongodb only stores milliseconds.
	return time.Now().UTC().Round(time.Millisecond)
}

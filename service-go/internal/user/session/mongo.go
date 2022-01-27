package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
)

const (
	batchLimit = 500
)

type Mongo struct {
	db *mongo.Database
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		db: db,
	}
}

func (m *Mongo) Create(ctx context.Context, requests ...Session) ([]Session, error) {
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
			return nil, fmt.Errorf("invalid request %w: %s", ErrValidation, err)
		}
		requests[i].CreatedAt = now
		docs[i] = requests[i]
	}

	_, err := m.db.Collection("sessions").InsertMany(ctx, docs)
	switch {
	case mongo.IsDuplicateKeyError(err):
		return nil, ErrDuplicatedEntry
	case err != nil:
		return nil, err
	}

	return requests, nil
}

func (m *Mongo) Fetch(ctx context.Context, lookup Lookup) ([]Session, error) {
	filter := bson.M{}
	if lookup.Key != "" {
		filter["_id"] = lookup.Key
	}
	if lookup.Owner != uuid.Nil {
		filter["owner"] = lookup.Owner
	}
	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}

	cur, err := m.db.Collection("sessions").Find(
		ctx,
		filter,
		&options.FindOptions{
			Limit: &lookup.Limit,
		},
	)
	if err != nil {
		return nil, err
	}
	var sessions []Session
	for cur.Next(ctx) {
		var s Session
		err := cur.Decode(&s)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (m *Mongo) FetchOne(ctx context.Context, lookup Lookup) (Session, error) {
	sessions, err := m.Fetch(ctx, lookup)
	if err != nil {
		return Session{}, err
	}
	if len(sessions) == 0 {
		return Session{}, ErrNotFound
	}
	if len(sessions) > 1 {
		return Session{}, ErrUnexpectedResult
	}
	return sessions[0], nil
}

func mongoNow() time.Time {
	// Mongodb only stores milliseconds.
	return time.Now().UTC().Round(time.Millisecond)
}

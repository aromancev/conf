package talk

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
	return &Mongo{db: db}
}

func (m *Mongo) Create(ctx context.Context, requests ...Talk) ([]Talk, error) {
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

	_, err := m.db.Collection("talks").InsertMany(ctx, docs)
	switch {
	case mongo.IsDuplicateKeyError(err):
		return nil, ErrDuplicateEntry
	case err != nil:
		return nil, err
	}

	return requests, nil
}

func (m *Mongo) Fetch(ctx context.Context, lookup Lookup) ([]Talk, error) {
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
	if lookup.Confa != uuid.Nil {
		filter["confa"] = lookup.Confa
	}
	if lookup.Handle != "" {
		filter["handle"] = lookup.Handle
	}
	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}

	cur, err := m.db.Collection("talks").Find(
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

	var talks []Talk
	for cur.Next(ctx) {
		var c Talk
		err := cur.Decode(&c)
		if err != nil {
			return nil, err
		}
		talks = append(talks, c)
	}

	return talks, nil
}

func (m *Mongo) FetchOne(ctx context.Context, lookup Lookup) (Talk, error) {
	talks, err := m.Fetch(ctx, lookup)
	if err != nil {
		return Talk{}, err
	}
	if len(talks) == 0 {
		return Talk{}, ErrNotFound
	}
	if len(talks) > 1 {
		return Talk{}, ErrUnexpectedResult
	}
	return talks[0], nil
}

func mongoNow() time.Time {
	// Mongodb only stores milliseconds.
	return time.Now().UTC().Round(time.Millisecond)
}

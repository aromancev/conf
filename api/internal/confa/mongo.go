package confa

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

func (m *Mongo) Create(ctx context.Context, requests ...Confa) ([]Confa, error) {
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

	_, err := m.db.Collection("confas").InsertMany(ctx, docs)
	switch {
	case mongo.IsDuplicateKeyError(err):
		return nil, ErrDuplicateEntry
	case err != nil:
		return nil, err
	}

	return requests, nil
}

func (m *Mongo) Fetch(ctx context.Context, lookup Lookup) ([]Confa, error) {
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
	if lookup.Handle != "" {
		filter["handle"] = lookup.Handle
	}
	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}

	cur, err := m.db.Collection("confas").Find(
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

	var confas []Confa
	for cur.Next(ctx) {
		var c Confa
		err := cur.Decode(&c)
		if err != nil {
			return nil, err
		}
		confas = append(confas, c)
	}

	return confas, nil
}

func (m *Mongo) FetchOne(ctx context.Context, lookup Lookup) (Confa, error) {
	confas, err := m.Fetch(ctx, lookup)
	if err != nil {
		return Confa{}, err
	}
	if len(confas) == 0 {
		return Confa{}, ErrNotFound
	}
	if len(confas) > 1 {
		return Confa{}, ErrUnexpectedResult
	}
	return confas[0], nil
}

func mongoNow() time.Time {
	// Mongodb only stores milliseconds.
	return time.Now().UTC().Round(time.Millisecond)
}
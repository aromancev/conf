package confa

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	batchLimit = 500
	collection = "confas"
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

	_, err := m.db.Collection(collection).InsertMany(ctx, docs)
	switch {
	case mongo.IsDuplicateKeyError(err):
		return nil, ErrDuplicateEntry
	case err != nil:
		return nil, err
	}

	return requests, nil
}

func (m *Mongo) Update(ctx context.Context, lookup Lookup, request Mask) (UpdateResult, error) {
	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}

	if err := request.Validate(); err != nil {
		return UpdateResult{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	update := bson.M{
		"$set": request,
	}
	res, err := m.db.Collection(collection).UpdateMany(ctx, lookup.Filter(), update)
	switch {
	case mongo.IsDuplicateKeyError(err):
		return UpdateResult{}, ErrDuplicateEntry
	case err != nil:
		return UpdateResult{}, err
	}
	return UpdateResult{Updated: res.ModifiedCount}, nil
}

func (m *Mongo) UpdateOne(ctx context.Context, lookup Lookup, request Mask) (Confa, error) {
	update := bson.M{
		"$set": request,
	}
	ret := options.After
	res := m.db.Collection(collection).FindOneAndUpdate(ctx, lookup.Filter(), update, &options.FindOneAndUpdateOptions{
		ReturnDocument: &ret,
	})
	switch {
	case errors.Is(res.Err(), mongo.ErrNoDocuments):
		return Confa{}, ErrNotFound
	case mongo.IsDuplicateKeyError(res.Err()):
		return Confa{}, ErrDuplicateEntry
	case res.Err() != nil:
		return Confa{}, res.Err()
	}

	var c Confa
	err := res.Decode(&c)
	if err != nil {
		return Confa{}, err
	}
	return c, nil
}

func (m *Mongo) Fetch(ctx context.Context, lookup Lookup) ([]Confa, error) {
	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}

	cur, err := m.db.Collection(collection).Find(
		ctx,
		lookup.Filter(),
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

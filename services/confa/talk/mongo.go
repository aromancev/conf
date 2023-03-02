package talk

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
	collection = "talks"
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
		if err := r.ValidateAtRest(); err != nil {
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

func (m *Mongo) UpdateOne(ctx context.Context, lookup Lookup, request Update) (Talk, error) {
	update := bson.M{
		"$set": request,
	}
	ret := options.After
	res := m.db.Collection(collection).FindOneAndUpdate(ctx, mongoFilter(lookup), update, &options.FindOneAndUpdateOptions{
		ReturnDocument: &ret,
	})
	switch {
	case errors.Is(res.Err(), mongo.ErrNoDocuments):
		return Talk{}, ErrNotFound
	case mongo.IsDuplicateKeyError(res.Err()):
		return Talk{}, ErrDuplicateEntry
	case res.Err() != nil:
		return Talk{}, res.Err()
	}

	var t Talk
	err := res.Decode(&t)
	if err != nil {
		return Talk{}, err
	}
	return t, nil
}

func (m *Mongo) Fetch(ctx context.Context, lookup Lookup) ([]Talk, error) {
	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}

	order := -1
	if lookup.Asc {
		order = 1
	}

	cur, err := m.db.Collection(collection).Find(
		ctx,
		mongoFilter(lookup),
		&options.FindOptions{
			Sort: bson.D{
				{Key: "createdAt", Value: order},
				{Key: "_id", Value: order},
			},
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
		return Talk{}, ErrAmbigiousLookup
	}
	return talks[0], nil
}

func mongoNow() time.Time {
	// Mongodb only stores milliseconds.
	return time.Now().UTC().Round(time.Millisecond)
}

func mongoFilter(l Lookup) bson.M {
	filter := make(bson.M)
	orderComp := "$lt"
	if l.Asc {
		orderComp = "$gt"
	}
	switch {
	case l.ID != uuid.Nil:
		filter["_id"] = l.ID
	case l.Handle != "":
		filter["handle"] = l.Handle
	case l.Room != uuid.Nil:
		filter["roomId"] = l.Room
	case !l.From.CreatedAt.IsZero() && l.From.ID != uuid.Nil:
		filter["$or"] = bson.A{
			bson.M{
				"createdAt": bson.M{
					orderComp: l.From.CreatedAt,
				},
			},
			bson.M{
				"createdAt": l.From.CreatedAt,
				"_id": bson.M{
					orderComp: l.From.ID,
				},
			},
		}
	case !l.From.CreatedAt.IsZero():
		filter["createdAt"] = bson.M{
			orderComp: l.From.CreatedAt,
		}
	case l.From.ID != uuid.Nil:
		filter["_id"] = bson.M{
			orderComp: l.From.ID,
		}
	}
	if l.Owner != uuid.Nil {
		filter["ownerId"] = l.Owner
	}
	if l.Confa != uuid.Nil {
		filter["confaId"] = l.Confa
	}
	if len(l.StateIn) != 0 {
		filter["state"] = bson.M{
			"$in": l.StateIn,
		}
	}
	return filter
}

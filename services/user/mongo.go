package user

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
	return &Mongo{db: db}
}

func (m *Mongo) Create(ctx context.Context, requests ...User) ([]User, error) {
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
		for j := range r.Idents {
			requests[i].Idents[j].CreatedAt = now
		}
		requests[i].CreatedAt = now
		docs[i] = requests[i]
	}

	_, err := m.db.Collection("users").InsertMany(ctx, docs)
	switch {
	case mongo.IsDuplicateKeyError(err):
		return nil, ErrDuplicatedEntry
	case err != nil:
		return nil, err
	}

	return requests, nil
}

func (m *Mongo) GetOrCreate(ctx context.Context, request User) (User, error) {
	now := mongoNow()
	if err := request.Validate(); err != nil {
		return User{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	for i := range request.Idents {
		request.Idents[i].CreatedAt = now
	}
	request.CreatedAt = now

	var match bson.A
	for _, ident := range request.Idents {
		match = append(match, bson.M{
			"platform": ident.Platform,
			"value":    ident.Value,
		})
	}
	res := m.db.Collection("users").FindOneAndUpdate(
		ctx,
		bson.M{
			"idents": bson.M{
				"$elemMatch": bson.M{
					"$or": match,
				},
			},
		},
		bson.M{
			"$setOnInsert": request,
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	)
	if err := res.Err(); err != nil {
		return User{}, err
	}
	if err := res.Decode(&request); err != nil {
		return User{}, err
	}

	return request, nil
}

func (m *Mongo) Fetch(ctx context.Context, lookup Lookup) ([]User, error) {
	filter := bson.M{}
	if lookup.ID != uuid.Nil {
		filter["_id"] = lookup.ID
	}

	cur, err := m.db.Collection("users").Find(
		ctx,
		filter,
		&options.FindOptions{
			Limit: &lookup.Limit,
		},
	)
	if err != nil {
		return nil, err
	}
	var users []User
	for cur.Next(ctx) {
		var u User
		err := cur.Decode(&u)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (m *Mongo) FetchOne(ctx context.Context, lookup Lookup) (User, error) {
	users, err := m.Fetch(ctx, lookup)
	if err != nil {
		return User{}, err
	}
	if len(users) == 0 {
		return User{}, ErrNotFound
	}
	if len(users) > 1 {
		return User{}, ErrUnexpectedResult
	}
	return users[0], nil
}

func mongoNow() time.Time {
	// Mongodb only stores milliseconds.
	return time.Now().UTC().Round(time.Millisecond)
}

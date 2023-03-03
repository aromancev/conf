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
	collection = "users"
)

type Mongo struct {
	db *mongo.Database
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{db: db}
}

func (m *Mongo) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	sess, err := m.db.Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer sess.EndSession(ctx)

	_, err = sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})
	return err
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
		for j := range r.Idents {
			requests[i].Idents[j] = requests[i].Idents[j].Normalized()
		}
		requests[i].CreatedAt = now
		docs[i] = requests[i]
		if err := r.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrValidation, err)
		}
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

func (m *Mongo) GetOrCreate(ctx context.Context, request User) (User, error) {
	now := mongoNow()
	for i := range request.Idents {
		request.Idents[i] = request.Idents[i].Normalized()
	}
	request.CreatedAt = now
	if err := request.Validate(); err != nil {
		return User{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}

	var match bson.A
	for _, ident := range request.Idents {
		match = append(match, bson.M{
			"platform": ident.Platform,
			"value":    ident.Value,
		})
	}
	res := m.db.Collection(collection).FindOneAndUpdate(
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

func (m *Mongo) Update(ctx context.Context, lookup Lookup, request Update) (UpdateResult, error) {
	if err := request.Validate(); err != nil {
		return UpdateResult{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	update := bson.M{
		"$set": request,
	}
	res, err := m.db.Collection(collection).UpdateMany(ctx, mongoFilter(lookup), update)
	switch {
	case mongo.IsDuplicateKeyError(err):
		return UpdateResult{}, ErrDuplicateEntry
	case err != nil:
		return UpdateResult{}, err
	}
	return UpdateResult{Updated: res.ModifiedCount}, nil
}

func (m *Mongo) UpdateOne(ctx context.Context, lookup Lookup, request Update) (User, error) {
	if err := request.Validate(); err != nil {
		return User{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	update := bson.M{
		"$set": request,
	}
	returnAfter := options.After
	res := m.db.Collection(collection).FindOneAndUpdate(
		ctx,
		mongoFilter(lookup),
		update,
		&options.FindOneAndUpdateOptions{
			ReturnDocument: &returnAfter,
		},
	)
	switch {
	case mongo.IsDuplicateKeyError(res.Err()):
		return User{}, ErrDuplicateEntry
	case errors.Is(res.Err(), mongo.ErrNoDocuments):
		return User{}, ErrNotFound
	case res.Err() != nil:
		return User{}, res.Err()
	}

	var user User
	err := res.Decode(&user)
	if err != nil {
		return User{}, fmt.Errorf("faile to decode user: %w", err)
	}
	return user, nil
}

func (m *Mongo) Fetch(ctx context.Context, lookup Lookup) ([]User, error) {
	if err := lookup.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidation, err)
	}

	cur, err := m.db.Collection(collection).Find(
		ctx,
		mongoFilter(lookup),
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

func mongoFilter(l Lookup) bson.M {
	filter := make(bson.M)
	if l.ID != uuid.Nil {
		filter["_id"] = l.ID
	}
	if len(l.Idents) > 0 {
		var match bson.A
		for _, ident := range l.Idents {
			ident = ident.Normalized()
			match = append(match, bson.M{
				"platform": ident.Platform,
				"value":    ident.Value,
			})
		}
		filter["idents"] = bson.M{
			"$elemMatch": bson.M{
				"$or": match,
			},
		}
	}
	if l.WithoutPassword {
		filter["passwordHash"] = bson.M{
			"$exists": false,
		}
	}
	return filter
}

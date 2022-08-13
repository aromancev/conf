package record

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	batchLimit = 500
	collection = "records"
)

type UpdateResult struct {
	ModifiedCount int64
}

type Mongo struct {
	db *mongo.Database
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		db: db,
	}
}

func (m *Mongo) FetchOrStart(ctx context.Context, record Record) (Record, error) {
	if err := record.Validate(); err != nil {
		return Record{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}

	filter := bson.M{
		"active": true,
		"room":   record.Room,
	}
	if record.Key != "" {
		filter["key"] = record.Key
	} else {
		record.Key = record.ID.String()
	}
	now := mongoNow()
	record.CreatedAt = now
	record.StartedAt = now
	record.Active = true

	res := m.db.Collection(collection).FindOneAndUpdate(
		ctx,
		filter,
		bson.M{
			"$setOnInsert": record,
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	)
	err := res.Err()
	switch {
	case mongo.IsDuplicateKeyError(err):
		return Record{}, ErrDuplicateEntry
	case err != nil:
		return Record{}, err
	}
	if err := res.Decode(&record); err != nil {
		return Record{}, err
	}
	return record, nil
}

func (m *Mongo) Stop(ctx context.Context, lookup Lookup) (UpdateResult, error) {
	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}
	update := bson.M{
		"$set": bson.M{
			"stoppedAt": time.Now(),
		},
		"$unset": bson.M{
			"active": nil,
		},
	}

	filter := mongoFilter(lookup)
	filter["active"] = true // Should only stop if active.

	res, err := m.db.Collection(collection).UpdateMany(
		ctx,
		filter,
		update,
	)
	switch {
	case mongo.IsDuplicateKeyError(err):
		return UpdateResult{}, ErrDuplicateEntry
	case err != nil:
		return UpdateResult{}, err
	}
	return UpdateResult{
		ModifiedCount: res.ModifiedCount,
	}, nil
}

func (m *Mongo) Fetch(ctx context.Context, lookup Lookup) ([]Record, error) {
	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}

	cur, err := m.db.Collection(collection).Find(
		ctx,
		mongoFilter(lookup),
		&options.FindOptions{
			Sort:  bson.M{"key": 1},
			Limit: &lookup.Limit,
		},
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var records []Record
	for cur.Next(ctx) {
		var r Record
		err := cur.Decode(&r)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

func (m *Mongo) FetchOne(ctx context.Context, lookup Lookup) (Record, error) {
	confas, err := m.Fetch(ctx, lookup)
	if err != nil {
		return Record{}, err
	}
	if len(confas) == 0 {
		return Record{}, ErrNotFound
	}
	if len(confas) > 1 {
		return Record{}, ErrAmbigiousLookup
	}
	return confas[0], nil
}

func mongoNow() time.Time {
	// Mongodb only stores milliseconds.
	return time.Now().UTC().Round(time.Millisecond)
}

func mongoFilter(l Lookup) bson.M {
	filter := make(bson.M)
	switch {
	case l.ID != uuid.Nil:
		filter["_id"] = l.ID
	case l.FromKey != "":
		filter["key"] = bson.M{
			"$gt": l.FromKey,
		}
	}
	if l.Room != uuid.Nil {
		filter["roomId"] = l.Room
	}
	if l.Key != "" {
		filter["key"] = l.Key
	}
	return filter
}

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

func (m *Mongo) FetchOrStart(ctx context.Context, record Recording) (Recording, error) {
	if err := record.Validate(); err != nil {
		return Recording{}, fmt.Errorf("%w: %s", ErrValidation, err)
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
		return Recording{}, ErrDuplicateEntry
	case err != nil:
		return Recording{}, err
	}
	if err := res.Decode(&record); err != nil {
		return Recording{}, err
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

func (m *Mongo) UpdateRecords(ctx context.Context, lookup Lookup, records Records) (Recording, error) {
	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}

	if records.IsZero() {
		return Recording{}, fmt.Errorf("%w: records should not be empty", ErrValidation)
	}

	addToSet := make(bson.M)
	if len(records.RecordingStarted) != 0 {
		addToSet["records.recordingStarted"] = bson.M{
			"$each": records.RecordingStarted,
		}
	}
	if len(records.RecordingFinished) != 0 {
		addToSet["records.recordingFinished"] = bson.M{
			"$each": records.RecordingFinished,
		}
	}
	if len(records.ProcessingStarted) != 0 {
		addToSet["records.processingStarted"] = bson.M{
			"$each": records.ProcessingStarted,
		}
	}
	if len(records.ProcessingFinished) != 0 {
		addToSet["records.processingFinished"] = bson.M{
			"$each": records.ProcessingFinished,
		}
	}

	ret := options.After
	res := m.db.Collection(collection).FindOneAndUpdate(
		ctx,
		mongoFilter(lookup),
		bson.M{
			"$addToSet": addToSet,
		},
		&options.FindOneAndUpdateOptions{
			ReturnDocument: &ret,
		},
	)
	err := res.Err()
	if err != nil {
		return Recording{}, fmt.Errorf("failed to update recording: %w", err)
	}
	var r Recording
	err = res.Decode(&r)
	if err != nil {
		return Recording{}, err
	}
	return r, nil
}

func (m *Mongo) Fetch(ctx context.Context, lookup Lookup) ([]Recording, error) {
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

	var records []Recording
	for cur.Next(ctx) {
		var r Recording
		err := cur.Decode(&r)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

func (m *Mongo) FetchOne(ctx context.Context, lookup Lookup) (Recording, error) {
	confas, err := m.Fetch(ctx, lookup)
	if err != nil {
		return Recording{}, err
	}
	if len(confas) == 0 {
		return Recording{}, ErrNotFound
	}
	if len(confas) > 1 {
		return Recording{}, ErrAmbigiousLookup
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

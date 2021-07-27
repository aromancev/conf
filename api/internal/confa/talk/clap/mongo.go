package clap

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	db *mongo.Database
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		db: db,
	}
}

func (m *Mongo) CreateOrUpdate(ctx context.Context, request Clap) (uuid.UUID, error) {
	err := request.Validate()
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %s", ErrValidation, err)
	}

	res, err := m.db.Collection("claps").UpdateOne(
		ctx,
		bson.M{
			"confaId":   request.Confa,
			"ownerId":   request.Owner,
			"speakerId": request.Speaker,
			"talkId":    request.Talk,
		},
		bson.M{
			"$set": bson.M{
				"confaId":   request.Confa,
				"ownerId":   request.Owner,
				"speakerId": request.Speaker,
				"talkId":    request.Talk,
				"value":     request.Value,
			},
			"$setOnInsert": bson.M{
				"_id": request.ID,
			},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return uuid.Nil, err
	}
	if res.UpsertedCount == 0 {
		return request.ID, nil
	}
	bin, ok := res.UpsertedID.(primitive.Binary)
	if !ok || len(bin.Data) != 16 { // binary UUID length = 16.
		return uuid.Nil, errors.New("failed to parse upserted id")
	}
	var id uuid.UUID
	copy(id[:], bin.Data)
	return id, err
}

func (m *Mongo) Aggregate(ctx context.Context, lookup Lookup) (uint64, error) {
	var match bson.D
	if lookup.Confa != uuid.Nil {
		match = append(match, bson.E{Key: "confaId", Value: lookup.Confa})
	}
	if lookup.Speaker != uuid.Nil {
		match = append(match, bson.E{Key: "speakerId", Value: lookup.Speaker})
	}
	if lookup.Talk != uuid.Nil {
		match = append(match, bson.E{Key: "talkId", Value: lookup.Talk})
	}
	group := bson.M{
		"_id": "",
		"sum": bson.M{"$sum": "$value"},
	}
	cur, err := m.db.Collection("claps").Aggregate(
		ctx,
		mongo.Pipeline{
			{{Key: "$match", Value: match}},
			{{Key: "$group", Value: group}},
		},
	)
	if err != nil {
		return 0, err
	}
	defer cur.Close(ctx)

	if !cur.Next(ctx) {
		return 0, nil
	}
	var claps struct {
		Sum uint64 `bson:"sum"`
	}
	err = cur.Decode(&claps)
	if err != nil {
		return 0, err
	}
	return claps.Sum, nil
}

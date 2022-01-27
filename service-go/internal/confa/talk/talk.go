package talk

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrValidation       = errors.New("invalid talk")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrDuplicateEntry   = errors.New("talk already exits")
	ErrPermissionDenied = errors.New("permission denied")
)

type Talk struct {
	ID          uuid.UUID `bson:"_id"`
	Confa       uuid.UUID `bson:"confaId"`
	Owner       uuid.UUID `bson:"ownerId"`
	Speaker     uuid.UUID `bson:"speakerId"`
	Room        uuid.UUID `bson:"roomId"`
	Handle      string    `bson:"handle"`
	Title       string    `bson:"title"`
	Description string    `bson:"description"`
	CreatedAt   time.Time `bson:"createdAt"`
}

var validHandle = regexp.MustCompile("^[a-z0-9-]{4,64}$")
var validTitle = regexp.MustCompile("^[a-zA-Z0-9- ]{0,64}$")

func (t Talk) Validate() error {
	if !validHandle.MatchString(t.Handle) {
		return errors.New("invalid handle")
	}
	return nil
}

func (t Talk) ValidateAtRest() error {
	if t.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if t.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	if t.Speaker == uuid.Nil {
		return errors.New("speaker should not be empty")
	}
	if t.Confa == uuid.Nil {
		return errors.New("confa should not be empty")
	}
	if t.Room == uuid.Nil {
		return errors.New("room should not be empty")
	}
	if !validHandle.MatchString(t.Handle) {
		return errors.New("invalid handle")
	}
	return nil
}

type Mask struct {
	Handle      *string `bson:"handle,omitempty"`
	Title       *string `bson:"title,omitempty"`
	Description *string `bson:"description,omitempty"`
}

func (m Mask) Validate() error {
	if m.Handle == nil && m.Title == nil && m.Description == nil {
		return errors.New("no fields provided")
	}
	if m.Handle != nil && !validHandle.MatchString(*m.Handle) {
		return errors.New("invalid handle")
	}
	if m.Title != nil && !validTitle.MatchString(*m.Title) {
		return errors.New("invalid title")
	}
	if m.Description != nil && len(*m.Description) > maxDescription {
		return errors.New("ivalid description")
	}
	return nil
}

type Lookup struct {
	ID      uuid.UUID
	Owner   uuid.UUID
	Confa   uuid.UUID
	Speaker uuid.UUID
	Handle  string
	Limit   int64
	From    uuid.UUID
}

func (l Lookup) Filter() bson.M {
	filter := make(bson.M)
	switch {
	case l.ID != uuid.Nil:
		filter["_id"] = l.ID
	case l.From != uuid.Nil:
		filter["_id"] = bson.M{
			"$gt": l.From,
		}
	}
	if l.Owner != uuid.Nil {
		filter["ownerId"] = l.Owner
	}
	if l.Confa != uuid.Nil {
		filter["confaId"] = l.Confa
	}
	if l.Handle != "" {
		filter["handle"] = l.Handle
	}
	return filter
}

const (
	maxDescription = 5000
)

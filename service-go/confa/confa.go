package confa

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrValidation       = errors.New("invalid confa")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrDuplicateEntry   = errors.New("confa already exists")
)

type Confa struct {
	ID          uuid.UUID `bson:"_id"`
	Owner       uuid.UUID `bson:"ownerId"`
	Handle      string    `bson:"handle"`
	Title       string    `bson:"title,omitempty"`
	Description string    `bson:"description,omitempty"`
	CreatedAt   time.Time `bson:"createdAt"`
}

var validHandle = regexp.MustCompile("^[a-z0-9-]{4,64}$")
var validTitle = regexp.MustCompile("^[a-zA-Z0-9- ]{0,64}$")

func (c Confa) Validate() error {
	if c.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if !validHandle.MatchString(c.Handle) {
		return errors.New("invalid handle")
	}
	if !validTitle.MatchString(c.Title) {
		return errors.New("invalid title")
	}
	if len(c.Description) > maxDescription {
		return errors.New("ivalid description")
	}
	if c.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
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
	ID     uuid.UUID
	Owner  uuid.UUID
	Handle string
	Limit  int64
	From   uuid.UUID
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
	if l.Handle != "" {
		filter["handle"] = l.Handle
	}
	return filter
}

type UpdateResult struct {
	Updated int64
}

const (
	maxDescription = 5000
)

package talk

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation       = errors.New("invalid talk")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrDuplicateEntry   = errors.New("talk already exits")
	ErrPermissionDenied = errors.New("permission denied")
)

type Talk struct {
	ID        uuid.UUID `bson:"_id"`
	Confa     uuid.UUID `bson:"confaId"`
	Owner     uuid.UUID `bson:"ownerId"`
	Speaker   uuid.UUID `bson:"speakerId"`
	Room      uuid.UUID `bson:"roomId"`
	Handle    string    `bson:"handle"`
	CreatedAt time.Time `bson:"createdAt"`
}

var validHandle = regexp.MustCompile("^[a-z0-9-]{1,64}$")

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

type Lookup struct {
	ID      uuid.UUID
	Owner   uuid.UUID
	Confa   uuid.UUID
	Speaker uuid.UUID
	Handle  string
	Limit   int64
	From    uuid.UUID
}

package room

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation       = errors.New("invalid room")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
)

type Room struct {
	ID        uuid.UUID `bson:"_id"`
	Owner     uuid.UUID `bson:"ownerId"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (r Room) Validate() error {
	if r.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if r.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	return nil
}

type Lookup struct {
	ID    uuid.UUID
	Owner uuid.UUID
	Limit int64
	From  uuid.UUID
}

package record

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation      = errors.New("invalid room")
	ErrNotFound        = errors.New("not found")
	ErrAmbigiousLookup = errors.New("ambigious lookup")
	ErrDuplicateEntry  = errors.New("record already exists")
)

type Record struct {
	ID        uuid.UUID `bson:"_id"`
	Room      uuid.UUID `bson:"roomId"`
	Key       string    `bson:"key"`
	Active    bool      `bson:"active,omitempty"`
	StartedAt time.Time `bson:"startedAt"`
	StoppedAt time.Time `bson:"stoppedAt,omitempty"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (r Record) Validate() error {
	if r.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if r.Room == uuid.Nil {
		return errors.New("room should not be empty")
	}
	return nil
}

type Lookup struct {
	ID      uuid.UUID
	Room    uuid.UUID
	Key     string
	Limit   int64
	FromKey string
}

package talk

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var (
	ErrValidation       = errors.New("invalid talk")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
)

type Talk struct {
	ID     uuid.UUID
	Confa  uuid.UUID
	Handle string

	CreatedAt time.Time
}

func (t Talk) Validate() error {
	if t.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if t.Confa == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	return nil
}

type Lookup struct {
	ID    uuid.UUID
	Confa uuid.UUID
}

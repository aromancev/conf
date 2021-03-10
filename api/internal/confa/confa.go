package confa

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation       = errors.New("invalid confa")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
)

type Confa struct {
	ID     uuid.UUID
	Owner  uuid.UUID
	Handle string

	CreatedAt time.Time
}

func (c Confa) Validate() error {
	if c.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if c.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	return nil
}

type Lookup struct {
	ID    uuid.UUID
	Owner uuid.UUID
}

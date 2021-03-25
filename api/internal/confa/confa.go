package confa

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation       = errors.New("invalid confa")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrDuplicatedEntry  = errors.New("duplicated entry")
)

type Confa struct {
	ID        uuid.UUID `json:"id"`
	Owner     uuid.UUID `json:"owner"`
	Handle    string    `json:"handle"`
	CreatedAt time.Time `json:"createdAt"`
}

var validHandle = regexp.MustCompile("^[A-z,0-9,-]{1,64}$")

func (c Confa) Validate() error {
	if c.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if !validHandle.MatchString(c.Handle) {
		return errors.New("invalid handle")
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

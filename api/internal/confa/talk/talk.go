package talk

import (
	"errors"
	"github.com/google/uuid"
	"regexp"
	"time"
)

var (
	ErrValidation       = errors.New("invalid talk")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrDuplicatedEntry  = errors.New("duplicated entry")
	ErrPermissionDenied = errors.New("permission denied")
)

type Talk struct {
	ID        uuid.UUID `json:"id"`
	Confa     uuid.UUID `json:"confa"`
	Owner     uuid.UUID `json:"owner"`
	Handle    string    `json:"handle"`
	CreatedAt time.Time `json:"createdAt"`
}

var validHandle = regexp.MustCompile("^[A-z,0-9,-]{1,64}$")

func (t Talk) Validate() error {
	if t.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if t.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	if !validHandle.MatchString(t.Handle) {
		return errors.New("invalid handle")
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

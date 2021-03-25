package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrValidation       = errors.New("invalid user")
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

func (u User) Validate() error {
	if u.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	return nil
}

type Lookup struct {
	ID uuid.UUID
}

package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
)

type User struct {
	ID        uuid.UUID
	CreatedAt time.Time
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

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
	ErrDuplicateEntry   = errors.New("confa already exists")
)

type Confa struct {
	ID        uuid.UUID `bson:"_id"`
	Owner     uuid.UUID `bson:"ownerId"`
	Handle    string    `bson:"handle"`
	CreatedAt time.Time `bson:"createdAt"`
}

var validHandle = regexp.MustCompile("^[A-z0-9-]{1,64}$")

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
	ID     uuid.UUID
	Owner  uuid.UUID
	Handle string
	Limit  int64
	From   uuid.UUID
}

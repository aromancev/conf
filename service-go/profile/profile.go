package profile

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation       = errors.New("invalid profile")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrDuplicateEntry   = errors.New("profile already exists")
)

type Profile struct {
	ID          uuid.UUID `bson:"_id"`
	Owner       uuid.UUID `bson:"ownerId"`
	Handle      string    `bson:"handle"`
	DisplayName string    `bson:"displayName"`
	CreatedAt   time.Time `bson:"createdAt"`
}

var validHandle = regexp.MustCompile("^[a-z0-9-]{4,64}$")
var validDisplayName = regexp.MustCompile("^[a-zA-Z- ]{0,64}$")

func (p Profile) Validate() error {
	if p.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if p.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	if !validHandle.MatchString(p.Handle) {
		return errors.New("invalid handle")
	}
	if !validDisplayName.MatchString(p.DisplayName) {
		return errors.New("invalid display name")
	}
	if p.CreatedAt.IsZero() {
		return errors.New("created at should not be empty")
	}
	return nil
}

type Lookup struct {
	ID     uuid.UUID
	Owners []uuid.UUID
	Handle string
	Limit  int64
	From   uuid.UUID
}

func (l Lookup) Validate() error {
	if len(l.Owners) > batchLimit {
		return errors.New("too many owners")
	}
	return nil
}

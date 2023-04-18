package confa

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation      = errors.New("invalid confa")
	ErrNotFound        = errors.New("not found")
	ErrAmbiguousLookup = errors.New("ambiguous lookup")
	ErrDuplicateEntry  = errors.New("confa already exists")
)

type Confa struct {
	ID          uuid.UUID `bson:"_id"`
	Owner       uuid.UUID `bson:"ownerId"`
	Handle      string    `bson:"handle"`
	Title       string    `bson:"title,omitempty"`
	Description string    `bson:"description,omitempty"`
	CreatedAt   time.Time `bson:"createdAt"`
}

var validHandle = regexp.MustCompile(`^[a-z0-9-]{4,64}$`)
var validTitle = regexp.MustCompile(`^[^ ]*[\p{L}0-9- ]{2,64}[^ ]*$`)

const (
	maxDescription = 5000
)

func (c Confa) Validate() error {
	if c.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if !validHandle.MatchString(c.Handle) {
		return errors.New("invalid handle")
	}
	if c.Title != "" && !validTitle.MatchString(c.Title) {
		return errors.New("invalid title")
	}
	if len(c.Description) > maxDescription {
		return errors.New("ivalid description")
	}
	if c.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	return nil
}

type Update struct {
	Handle      *string `bson:"handle,omitempty"`
	Title       *string `bson:"title,omitempty"`
	Description *string `bson:"description,omitempty"`
}

func (m Update) Validate() error {
	if m.Handle == nil && m.Title == nil && m.Description == nil {
		return errors.New("no fields provided")
	}
	if m.Handle != nil && !validHandle.MatchString(*m.Handle) {
		return errors.New("invalid handle")
	}
	if m.Title != nil && !validTitle.MatchString(*m.Title) {
		return errors.New("invalid title")
	}
	if m.Description != nil && len(*m.Description) > maxDescription {
		return errors.New("ivalid description")
	}
	return nil
}

type UpdateResult struct {
	Updated int64
}

type Lookup struct {
	ID     uuid.UUID
	Owner  uuid.UUID
	Handle string
	Limit  int64
	From   From
	Asc    bool
}

type From struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

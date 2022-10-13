package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrDuplicatedEntry  = errors.New("duplicated entry")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrValidation       = errors.New("invalid user")
)

type User struct {
	ID        uuid.UUID `bson:"_id"`
	CreatedAt time.Time `bson:"createdAt"`
	Idents    []Ident   `bson:"idents"`
}

func (u User) Validate() error {
	if u.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if len(u.Idents) > 10 {
		return errors.New("too many identifiers")
	}
	idents := make(map[string]struct{}, len(u.Idents))
	for _, ident := range u.Idents {
		if err := ident.Validate(); err != nil {
			return fmt.Errorf("identifier is not valid: %w", err)
		}
		idx := string(ident.Platform) + ident.Value
		if _, ok := idents[idx]; ok {
			return fmt.Errorf("identifier is not valid: %w", ErrDuplicatedEntry)
		}
		idents[idx] = struct{}{}
	}
	return nil
}

type Platform string

const (
	PlatformUnknown Platform = ""
	PlatformEmail   Platform = "email"
	PlatformTwitter Platform = "twitter"
	PlatformGithub  Platform = "github"
)

func (p Platform) Validate() error {
	switch p {
	case PlatformEmail, PlatformTwitter, PlatformGithub:
		return nil
	}
	return errors.New("unknown platform")
}

type Ident struct {
	ID        uuid.UUID `bson:"_id" json:"id"`
	Platform  Platform  `bson:"platform" json:"platform"`
	Value     string    `bson:"value" json:"value"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

func (i Ident) Validate() error {
	if i.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if err := i.Platform.Validate(); err != nil {
		return err
	}
	if i.Value == "" {
		return errors.New("value should not be empty")
	}
	return nil
}

type Lookup struct {
	ID    uuid.UUID
	Limit int64
}

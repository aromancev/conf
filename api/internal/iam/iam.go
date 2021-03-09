package iam

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation       = errors.New("invalid object")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrDuplicatedEntry  = errors.New("duplicated entry")
)

type Platform string

const (
	PlatformUnknown Platform = ""
	PlatformEmail   Platform = "email"
)

func (p Platform) Validate() error {
	switch p {
	case PlatformEmail:
		return nil
	}
	return errors.New("unknown platform")
}

type Ident struct {
	ID        uuid.UUID
	Owner     uuid.UUID
	Platform  Platform
	Value     string
	CreatedAt time.Time
}

func (i Ident) Validate() error {
	if i.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if i.Owner == uuid.Nil {
		return errors.New("user should not be empty")
	}
	if err := i.Platform.Validate(); err != nil {
		return err
	}
	if i.Value == "" {
		return errors.New("value should not be empty")
	}
	return nil
}

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

type IdentLookup struct {
	ID       uuid.UUID
	Owner    uuid.UUID
	Matching []Ident
}

func Authenticate(r *http.Request) (User, error) {
	return User{
		ID: uuid.New(),
	}, nil
}

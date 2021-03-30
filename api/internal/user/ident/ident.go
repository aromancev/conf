package ident

import (
	"errors"
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
	ID        uuid.UUID `json:"id"`
	Owner     uuid.UUID `json:"owner"`
	Platform  Platform  `json:"platform"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
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

type Lookup struct {
	ID       uuid.UUID
	Owner    uuid.UUID
	Matching []Ident
}

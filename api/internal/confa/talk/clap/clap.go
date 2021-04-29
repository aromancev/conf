package clap

import (
	"errors"
	"github.com/google/uuid"
)


var (
	ErrDuplicatedEntry  = errors.New("duplicated entry")
	ErrValidation       = errors.New("invalid clap")
)

type Clap struct {
	ID        uuid.UUID `json:"id"`
	Confa     uuid.UUID `json:"confa"`
	Owner     uuid.UUID `json:"owner"`
	Speaker   uuid.UUID `json:"speaker"`
	Talk      uuid.UUID `json:"talk"`
	Claps  	  int8 `json:"claps"`
}

func (c Clap) Validate() error {
	if c.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if c.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	if c.Speaker == uuid.Nil {
		return errors.New("speaker should not be empty")
	}
	if c.Confa == uuid.Nil {
		return errors.New("confa should not be empty")
	}
	if c.Talk == uuid.Nil {
		return errors.New("talk should not be empty")
	}
	if c.Claps > 50 {
		return errors.New("claps cannot be more than 50")
	}
	return nil
}


type Lookup struct {
	Confa uuid.UUID
	Speaker uuid.UUID
	Talk uuid.UUID
}

package clap

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrValidation = errors.New("invalid clap")
)

type Clap struct {
	ID      uuid.UUID `bson:"_id"`
	Confa   uuid.UUID `bson:"confaId"`
	Owner   uuid.UUID `bson:"ownerId"`
	Talk    uuid.UUID `bson:"talkId"`
	Speaker uuid.UUID `bson:"speakerId"`
	Value   uint      `bson:"value"`
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
	if c.Value > 50 {
		return errors.New("value cannot be more than 50")
	}
	return nil
}

type Lookup struct {
	Confa   uuid.UUID
	Speaker uuid.UUID
	Talk    uuid.UUID
}

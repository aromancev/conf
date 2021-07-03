package session

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation       = errors.New("invalid session")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrDuplicatedEntry  = errors.New("duplicated entry")
)

type Session struct {
	Key       string    `bson:"_id" json:"key"`
	Owner     uuid.UUID `bson:"owner" json:"owner"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

func NewKey() string {
	const keyLength = 96

	b := make([]byte, keyLength)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	key := base64.StdEncoding.EncodeToString(b)

	return key
}

func (s Session) Validate() error {
	if s.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	if s.Key == "" {
		return errors.New("session key should not be empty")
	}
	return nil
}

type Lookup struct {
	Key   string
	Owner uuid.UUID
	Limit int64
}

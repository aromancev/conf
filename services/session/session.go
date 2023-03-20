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
	Key       string    `bson:"_id"`
	Owner     uuid.UUID `bson:"owner"`
	CreatedAt time.Time `bson:"createdAt"`
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

func (l Lookup) Validate() error {
	if l.Key == "" && l.Owner == uuid.Nil {
		return errors.New("empty lookup not allowed")
	}
	return nil
}

type UpdateResult struct {
	Updated int64
}

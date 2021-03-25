package session

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/google/uuid"
	"time"
)

var (
	ErrValidation       = errors.New("invalid session")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrDuplicatedEntry  = errors.New("duplicated entry")
)

type Session struct {
	Key       string    `json:"key"`
	Owner     uuid.UUID `json:"owner"`
	CreatedAt time.Time `json:"createdAt"`
}

const byteKeyLength = 96 // you need 4*(n/3) chars to represent n bytes
func generateKey() string {
	b := make([]byte, byteKeyLength)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	key := base64.StdEncoding.EncodeToString(b)
	return key
}

func (s Session) GenerateKey() (Session, error) {
	if s.Key != "" {
		return s, errors.New("key is not empty")
	}

	key := generateKey()
	s.Key = key
	return s, nil
}

func (s Session) Validate() error {
	if s.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	return nil
}

type Lookup struct {
	Key   string
	Owner uuid.UUID
}

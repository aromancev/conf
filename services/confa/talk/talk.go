package talk

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation       = errors.New("invalid talk")
	ErrNotFound         = errors.New("not found")
	ErrAmbigiousLookup  = errors.New("ambigious lookup")
	ErrDuplicateEntry   = errors.New("talk already exits")
	ErrPermissionDenied = errors.New("permission denied")
	ErrWrongState       = errors.New("wrong state")
)

type State string

const (
	StateCreated   State = "CREATED"
	StateLive      State = "LIVE"
	StateRecording State = "RECORDING"
	StateEnded     State = "ENDED"
)

func (s State) Validate() error {
	switch s {
	case StateCreated, StateLive, StateRecording, StateEnded:
	default:
		return fmt.Errorf("should be one of %v", []State{StateCreated, StateLive, StateRecording, StateEnded})
	}
	return nil
}

type Talk struct {
	ID          uuid.UUID `bson:"_id"`
	Confa       uuid.UUID `bson:"confaId"`
	Owner       uuid.UUID `bson:"ownerId"`
	Speaker     uuid.UUID `bson:"speakerId"`
	Room        uuid.UUID `bson:"roomId"`
	Handle      string    `bson:"handle"`
	Title       string    `bson:"title"`
	Description string    `bson:"description"`
	State       State     `bson:"state"`
	CreatedAt   time.Time `bson:"createdAt"`
}

var validHandle = regexp.MustCompile("^[a-z0-9-]{4,64}$")
var validTitle = regexp.MustCompile("^[a-zA-Z0-9- ]{0,64}$")

const (
	maxDescription = 5000
)

func (t Talk) Validate() error {
	if !validHandle.MatchString(t.Handle) {
		return errors.New("invalid handle")
	}
	return nil
}

func (t Talk) ValidateAtRest() error {
	if t.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if t.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	if t.Speaker == uuid.Nil {
		return errors.New("speaker should not be empty")
	}
	if t.Confa == uuid.Nil {
		return errors.New("confa should not be empty")
	}
	if t.Room == uuid.Nil {
		return errors.New("room should not be empty")
	}
	if t.Title != "" && !validTitle.MatchString(t.Title) {
		return errors.New("invalid title")
	}
	if !validHandle.MatchString(t.Handle) {
		return errors.New("invalid handle")
	}
	if err := t.State.Validate(); err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}
	return nil
}

type Update struct {
	Handle      *string `bson:"handle,omitempty"`
	Title       *string `bson:"title,omitempty"`
	Description *string `bson:"description,omitempty"`
	State       *State  `bson:"state,omitempty"`
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
	if m.State != nil {
		if err := m.State.Validate(); err != nil {
			return fmt.Errorf("invalid state: %w", err)
		}
	}
	return nil
}

type Lookup struct {
	ID      uuid.UUID
	Owner   uuid.UUID
	Confa   uuid.UUID
	Speaker uuid.UUID
	Room    uuid.UUID
	Handle  string
	Limit   int64
	StateIn []State
	From    From
	Asc     bool
}

type From struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

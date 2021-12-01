package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrValidation       = errors.New("invalid event")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrCursorClosed     = errors.New("cursor closed")
	ErrShuttingDown     = errors.New("shutting down")
	ErrUnknownEvent     = errors.New("unknown event")
	ErrDuplicatedEntry  = errors.New("duplicated entry")
)

type Type string

const (
	TypePeerState Type = "peer_state"
	TypeMessage   Type = "message"
)

type Event struct {
	ID        uuid.UUID `bson:"_id" json:"id"`
	Owner     uuid.UUID `bson:"ownerId" json:"ownerId"`
	Room      uuid.UUID `bson:"roomId" json:"roomId"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	Payload   Payload   `bson:"payload" json:"payload"`
}

func (e Event) ValidateAtRest() error {
	if e.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if e.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	if e.Room == uuid.Nil {
		return errors.New("room should not be empty")
	}
	return e.Payload.Validate()
}

type Validatable interface {
	Validate() error
}

type Payload struct {
	Type    Type        `bson:"type" json:"type"`
	Payload Validatable `bson:"payload" json:"payload"`
}

func (p Payload) Validate() error {
	switch p.Type {
	case TypePeerState:
		if _, ok := p.Payload.(PayloadPeerState); !ok {
			return fmt.Errorf("invalid payload for type: %s", p.Type)
		}
	case TypeMessage:
		if _, ok := p.Payload.(PayloadMessage); !ok {
			return fmt.Errorf("invalid payload for type: %s", p.Type)
		}
	default:
		return fmt.Errorf("invalid type: %s", p.Type)
	}
	return p.Payload.Validate()
}

func (p *Payload) UnmarshalJSON(b []byte) error {
	var raw struct {
		T Type            `json:"type"`
		P json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	switch raw.T {
	case TypePeerState:
		var pl PayloadPeerState
		if err := json.Unmarshal(raw.P, &pl); err != nil {
			return err
		}
		p.Payload = pl
	case TypeMessage:
		var pl PayloadMessage
		if err := json.Unmarshal(raw.P, &pl); err != nil {
			return err
		}
		p.Payload = pl
	default:
		return ErrUnknownEvent
	}
	p.Type = raw.T
	return nil
}

func (p *Payload) UnmarshalBSON(b []byte) error {
	var raw struct {
		T Type     `bson:"type"`
		P bson.Raw `bson:"payload"`
	}
	err := bson.Unmarshal(b, &raw)
	if err != nil {
		return err
	}
	switch raw.T {
	case TypePeerState:
		var pl PayloadPeerState
		if err := bson.Unmarshal(raw.P, &pl); err != nil {
			return err
		}
		p.Payload = pl
	case TypeMessage:
		var pl PayloadMessage
		if err := bson.Unmarshal(raw.P, &pl); err != nil {
			return err
		}
		p.Payload = pl
	default:
		return fmt.Errorf("%w: %s", ErrUnknownEvent, raw.T)
	}
	p.Type = raw.T
	return nil
}

type PayloadPeerState struct {
	Status PeerStatus       `bson:"status,omitempty" json:"status,omitempty"`
	Tracks map[string]Track `bson:"tracks,omitempty" json:"tracks,omitempty"`
}

func (p PayloadPeerState) Validate() error {
	if len(p.Tracks) > 3 {
		return errors.New("no more than 3 tracks allowed")
	}
	for id, t := range p.Tracks {
		if id == "" {
			return fmt.Errorf("invalid track: id cannot be zero")
		}
		if err := t.Validate(); err != nil {
			return fmt.Errorf("invalid track: %w", err)
		}
	}
	switch p.Status {
	case "", PeerJoined, PeerLeft:
		return nil
	default:
		return errors.New("unknown peer status")
	}
}

type PeerStatus string

const (
	PeerJoined PeerStatus = "joined"
	PeerLeft   PeerStatus = "left"
)

type Track struct {
	Hint string `bson:"hint,omitempty" json:"hint,omitempty"`
}

func (t Track) Validate() error {
	switch t.Hint {
	case HintCamera, HintScreen, HintUserAudio, HintDeviceAudio:
	default:
		return errors.New("invalid hint")
	}
	return nil
}

type TrackHint string

const (
	HintCamera      = "camera"
	HintScreen      = "screen"
	HintUserAudio   = "user_audio"
	HintDeviceAudio = "device_audio"
)

type PayloadMessage struct {
	Text string `bson:"text" json:"text"`
}

func (p PayloadMessage) Validate() error {
	return nil
}

type Lookup struct {
	ID    uuid.UUID
	Room  uuid.UUID
	Limit int64
	From  From
	Asc   bool
}

type From struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

type Watcher interface {
	Watch(ctx context.Context, roomID uuid.UUID) (Cursor, error)
}

type Cursor interface {
	Next(ctx context.Context) (Event, error)
	Close(ctx context.Context) error
}

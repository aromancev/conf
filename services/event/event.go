package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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

type Watcher interface {
	Watch(ctx context.Context, roomID uuid.UUID) (Cursor, error)
}

type Cursor interface {
	Next(ctx context.Context) (Event, error)
	Close(ctx context.Context) error
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

type Event struct {
	ID        uuid.UUID `bson:"_id"`
	Room      uuid.UUID `bson:"roomId"`
	CreatedAt time.Time `bson:"createdAt"`
	Payload   Payload   `bson:"payload"`
}

func (e Event) Validate() error {
	if e.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if e.Room == uuid.Nil {
		return errors.New("room should not be empty")
	}
	return e.Payload.Validate()
}

type Payload struct {
	PeerState      *PayloadPeerState      `bson:"peerState"`
	Message        *PayloadMessage        `bson:"message"`
	Recording      *PayloadRecording      `bson:"recording"`
	TrackRecording *PayloadTrackRecording `bson:"trackRecording"`
	Reaction       *PayloadReaction       `bson:"reaction"`
}

func (p Payload) Validate() error {
	var hasPayload bool
	if p.PeerState != nil {
		if err := p.PeerState.Validate(); err != nil {
			return err
		}
		hasPayload = true
	}
	if p.Message != nil {
		if err := p.Message.Validate(); err != nil {
			return err
		}
		hasPayload = true
	}
	if p.Recording != nil {
		if err := p.Recording.Validate(); err != nil {
			return err
		}
		hasPayload = true
	}
	if p.TrackRecording != nil {
		if err := p.TrackRecording.Validate(); err != nil {
			return err
		}
		hasPayload = true
	}
	if p.Reaction != nil {
		if err := p.Reaction.Validate(); err != nil {
			return err
		}
		hasPayload = true
	}
	if !hasPayload {
		return errors.New("payload must not be empty")
	}
	return nil
}

type PayloadPeerState struct {
	Peer      uuid.UUID  `bson:"peerId"`
	SessionID uuid.UUID  `bson:"sessionId"`
	Status    PeerStatus `bson:"status,omitempty"`
	Tracks    []Track    `bson:"tracks,omitempty"`
}

func (p PayloadPeerState) Validate() error {
	if p.Peer == uuid.Nil {
		return errors.New("peer must not be nil")
	}
	if p.SessionID == uuid.Nil {
		return errors.New("session must not be nil")
	}
	if len(p.Tracks) > 3 {
		return errors.New("no more than 3 tracks allowed")
	}
	for _, t := range p.Tracks {
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
	ID   string    `bson:"id"`
	Hint TrackHint `bson:"hint"`
}

func (t Track) Validate() error {
	if t.ID == "" {
		return errors.New("id must not be empty")
	}
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
	From uuid.UUID `bson:"fromId"`
	Text string    `bson:"text"`
}

func (p PayloadMessage) Validate() error {
	if p.From == uuid.Nil {
		return errors.New("from must not be nil")
	}
	if p.Text == "" {
		return errors.New("text must not be empty")
	}
	return nil
}

type RecordStatus string

const (
	RecordingStarted = "started"
	RecordingStopped = "stopped"
)

type PayloadRecording struct {
	Status RecordStatus `bson:"status"`
}

func (p PayloadRecording) Validate() error {
	switch p.Status {
	case RecordingStarted, RecordingStopped:
	default:
		return errors.New("invalid status")
	}
	return nil
}

type PayloadTrackRecording struct {
	ID      uuid.UUID `bson:"id"`
	TrackID string    `bson:"trackId"`
}

func (p PayloadTrackRecording) Validate() error {
	if p.ID == uuid.Nil {
		return errors.New("id should not be zero")
	}
	if p.TrackID == "" {
		return errors.New("track id should not be zero")
	}
	return nil
}

type PayloadReaction struct {
	From     uuid.UUID `bson:"fromId"`
	Reaction Reaction  `bson:"reaction"`
}

func (p PayloadReaction) Validate() error {
	if p.From == uuid.Nil {
		return errors.New("fromId should not be zero")
	}
	var hasReaction bool
	if p.Reaction.Clap != nil {
		hasReaction = true
	}
	if !hasReaction {
		return errors.New("reaction must not be empty")
	}
	return nil
}

type Reaction struct {
	Clap *ReactionClap `bson:"clap,omitempty"`
}

func (r Reaction) Validate() error {
	return nil
}

type ReactionClap struct {
	IsStarting bool `bson:"isStarting"`
}

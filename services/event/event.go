package event

import (
	"context"
	"errors"
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
		return errors.New("id should not be zero")
	}
	if e.Room == uuid.Nil {
		return errors.New("room should not be zero")
	}
	if e.CreatedAt.IsZero() {
		return errors.New("created at should not be zero")
	}
	return e.Payload.Validate()
}

type Payload struct {
	PeerState   *PayloadPeerState   `bson:"peerState"`
	Message     *PayloadMessage     `bson:"message"`
	Recording   *PayloadRecording   `bson:"recording"`
	TrackRecord *PayloadTrackRecord `bson:"trackRecording"`
	Reaction    *PayloadReaction    `bson:"reaction"`
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
	if p.TrackRecord != nil {
		if err := p.TrackRecord.Validate(); err != nil {
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
}

func (p PayloadPeerState) Validate() error {
	if p.Peer == uuid.Nil {
		return errors.New("peer must not be nil")
	}
	if p.SessionID == uuid.Nil {
		return errors.New("session must not be nil")
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
	PeerJoined PeerStatus = "JOINED"
	PeerLeft   PeerStatus = "LEFT"
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
	RecordingStarted = "STARTED"
	RecordingStopped = "STOPPED"
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

type TrackKind string

const (
	TrackKindAudio TrackKind = "AUDIO"
	TrackKindVideo TrackKind = "VIDEO"
)

type TrackSource string

const (
	TrackSourceUnknown     TrackSource = "UNKNOWN"
	TrackSourceCamera      TrackSource = "CAMERA"
	TrackSourceMicrophone  TrackSource = "MICROPHONE"
	TrackSourceScreen      TrackSource = "SCREEN"
	TrackSourceScreenAudio TrackSource = "SCREEN_AUDIO"
)

type PayloadTrackRecord struct {
	RecordID uuid.UUID   `bson:"recordId"`
	Kind     TrackKind   `bson:"trackKind"`
	Source   TrackSource `bson:"trackSource"`
}

func (p PayloadTrackRecord) Validate() error {
	if p.RecordID == uuid.Nil {
		return errors.New("id should not be zero")
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

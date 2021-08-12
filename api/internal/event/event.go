package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation       = errors.New("invalid room")
	ErrNotFound         = errors.New("not found")
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrCursorClosed     = errors.New("cursor closed")
	ErrShuttingDown     = errors.New("shutting down")
)

type Type string

const (
	TypePeerJoined = "peer_joined"
)

func (t Type) Validate() error {
	switch t { // nolint: gocritic
	case TypePeerJoined:
		return nil
	}
	return fmt.Errorf("invalid type: %s", t)
}

type Event struct {
	ID        uuid.UUID   `bson:"_id"`
	Owner     uuid.UUID   `bson:"ownerId"`
	Room      uuid.UUID   `bson:"roomId"`
	Type      Type        `bson:"type"`
	Payload   interface{} `bson:"payload"`
	CreatedAt time.Time   `bson:"createdAt"`
}

func (e Event) ValidateAtRest() error {
	if e.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if e.Owner == uuid.Nil {
		return errors.New("owner should not be empty")
	}
	return nil
}

type Lookup struct {
	ID    uuid.UUID
	Room  uuid.UUID
	Limit int64
	From  uuid.UUID
}

type Watcher interface {
	Watch(ctx context.Context, roomID uuid.UUID) (Cursor, error)
}

type Cursor interface {
	Next(ctx context.Context) (Event, error)
	Close(ctx context.Context) error
}

type Iter struct {
	cur Cursor
	ev  Event
	err error
}

func NewIter(cur Cursor) *Iter {
	return &Iter{
		cur: cur,
	}
}

func (i *Iter) Next(ctx context.Context) bool {
	i.ev, i.err = i.cur.Next(ctx)
	return i.err != nil
}

func (i *Iter) Range(ctx context.Context, f func(Event)) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		i.ev, i.err = i.cur.Next(ctx)
		if i.err != nil {
			return i.err
		}
		if f != nil {
			f(i.ev)
		}
	}
}

func (i *Iter) Event() Event {
	return i.ev
}

func (i *Iter) Err() error {
	return i.err
}

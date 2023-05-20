package peer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aromancev/confa/event"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var (
	ErrValidation     = errors.New("validation error")
	ErrClosed         = errors.New("connection closed")
	ErrUnknownMessage = errors.New("unknown message")
)

type EventEmitter interface {
	EmitEvent(context.Context, event.Event) error
}

type Peer struct {
	emitter                       EventEmitter
	userID, roomID, peerSessionID uuid.UUID
	events                        event.Cursor
}

func NewPeer(ctx context.Context, userID, roomID uuid.UUID, events event.Cursor, emitter EventEmitter) *Peer {
	p := &Peer{
		emitter:       emitter,
		events:        events,
		userID:        userID,
		roomID:        roomID,
		peerSessionID: uuid.New(),
	}
	_, err := p.emit(ctx, event.Payload{
		PeerState: &event.PayloadPeerState{
			Peer:      p.userID,
			SessionID: p.peerSessionID,
			Status:    event.PeerJoined,
		},
	})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to emit peer status event.")
	}
	return p
}

func (p *Peer) SessionID() uuid.UUID {
	return p.peerSessionID
}

func (p *Peer) SendMessage(ctx context.Context, text string) (event.Event, error) {
	msg := event.PayloadMessage{
		From: p.userID,
		Text: text,
	}
	if err := msg.Validate(); err != nil {
		return event.Event{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}

	return p.emit(ctx, event.Payload{
		Message: &msg,
	})
}

func (p *Peer) SendReaction(ctx context.Context, reaction event.Reaction) (event.Event, error) {
	if err := reaction.Validate(); err != nil {
		return event.Event{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	return p.emit(ctx, event.Payload{
		Reaction: &event.PayloadReaction{
			From:     p.userID,
			Reaction: reaction,
		},
	})
}

func (p *Peer) RecieveEvent(ctx context.Context) (event.Event, error) {
	ev, err := p.events.Next(ctx)
	switch {
	case errors.Is(err, event.ErrUnknownEvent):
		return event.Event{}, ErrUnknownMessage
	case errors.Is(err, event.ErrCursorClosed):
		return event.Event{}, ErrClosed
	}
	return ev, err
}

func (p *Peer) Close(ctx context.Context) {
	if err := p.events.Close(ctx); err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to close events.")
	}

	// Peer leave event must be sent even if context is cancelled.
	emitCtx, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	_, err := p.emit(emitCtx, event.Payload{
		PeerState: &event.PayloadPeerState{
			Peer:      p.userID,
			SessionID: p.peerSessionID,
			Status:    event.PeerLeft,
		},
	})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to emit peer status event.")
	}
}

func (p *Peer) emit(ctx context.Context, payload event.Payload) (event.Event, error) {
	ev := event.Event{
		ID:        uuid.New(),
		Room:      p.roomID,
		CreatedAt: time.Now(),
		Payload:   payload,
	}
	err := p.emitter.EmitEvent(ctx, ev)
	switch {
	case errors.Is(err, event.ErrValidation):
		return event.Event{}, ErrValidation
	case err != nil:
		return event.Event{}, err
	}
	return ev, nil
}

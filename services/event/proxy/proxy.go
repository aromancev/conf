package proxy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/internal/platform/signal"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
	"gortc.io/sdp"
)

var (
	ErrValidation     = errors.New("validation error")
	ErrClosed         = errors.New("connection closed")
	ErrUnknownMessage = errors.New("unknown message")
)

type Signal interface {
	Send(context.Context, signal.Message) error
	Receive(context.Context) (signal.Message, error)
}

type EventEmitter interface {
	EmitEvent(context.Context, event.Event) error
}

type State struct {
	Tracks []event.Track `json:"tracks"`
}

func (s State) Validate() error {
	if len(s.Tracks) > 3 {
		return errors.New("no more than 3 tracks allowed")
	}
	for _, t := range s.Tracks {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("invalid track: %w", err)
		}
	}
	return nil
}

func (s State) Track(id string) (event.Track, bool) {
	for _, t := range s.Tracks {
		if t.ID == id {
			return t, true
		}
	}
	return event.Track{}, false
}

type Proxy struct {
	emitter        EventEmitter
	userID, roomID uuid.UUID
	state          State
	events         event.Cursor
}

func NewProxy(ctx context.Context, userID, roomID uuid.UUID, events event.Cursor, emitter EventEmitter) *Proxy {
	p := &Proxy{
		emitter: emitter,
		events:  events,
		userID:  userID,
		roomID:  roomID,
	}
	_, err := p.emit(ctx, event.Payload{
		PeerState: &event.PayloadPeerState{
			Peer:   p.userID,
			Status: event.PeerJoined,
		},
	})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to emit peer status event.")
	}
	return p
}

func (p *Proxy) SendSignal(ctx context.Context, client Signal, msg signal.Message) error {
	switch {
	case msg.Join != nil:
		pl := msg.Join
		if pl.UserID != p.userID.String() {
			return fmt.Errorf("%w: %s", ErrValidation, "invalid user")
		}
		if pl.SessionID != p.roomID.String() {
			return fmt.Errorf("%w: %s", ErrValidation, "invalid room")
		}
		return client.Send(ctx, msg)
	case msg.Offer != nil:
		pl := msg.Offer
		tracks, err := p.tracks(pl.Description)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrValidation, err)
		}
		if err := client.Send(ctx, msg); err != nil {
			return err
		}
		_, err = p.emit(ctx, event.Payload{
			PeerState: &event.PayloadPeerState{
				Peer:   p.userID,
				Tracks: tracks,
			},
		})
		return err
	}

	return client.Send(ctx, msg)
}

func (p *Proxy) SendMessage(ctx context.Context, text string) (event.Event, error) {
	return p.emit(ctx, event.Payload{
		Message: &event.PayloadMessage{
			From: p.userID,
			Text: text,
		},
	})
}

func (p *Proxy) SendState(ctx context.Context, state State) (State, error) {
	if err := state.Validate(); err != nil {
		return State{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	p.state = state
	return state, nil
}

func (p *Proxy) RecieveEvent(ctx context.Context) (event.Event, error) {
	ev, err := p.events.Next(ctx)
	switch {
	case errors.Is(err, event.ErrUnknownEvent):
		return event.Event{}, ErrUnknownMessage
	case errors.Is(err, event.ErrCursorClosed):
		return event.Event{}, ErrClosed
	}
	return ev, err
}

func (p *Proxy) ReceiveSignal(ctx context.Context, client Signal) (signal.Message, error) {
	msg, err := client.Receive(ctx)
	switch {
	case errors.Is(err, signal.ErrUnknownMessage):
		return signal.Message{}, ErrUnknownMessage
	case errors.Is(err, signal.ErrClosed):
		return signal.Message{}, ErrClosed
	}
	return msg, err
}

func (p *Proxy) Close(ctx context.Context) {
	if err := p.events.Close(ctx); err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to close events.")
	}

	_, err := p.emit(ctx, event.Payload{
		PeerState: &event.PayloadPeerState{
			Peer:   p.userID,
			Status: event.PeerLeft,
		},
	})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to emit peer status event.")
	}
}

func (p *Proxy) emit(ctx context.Context, payload event.Payload) (event.Event, error) {
	ev := event.Event{
		ID:      uuid.New(),
		Room:    p.roomID,
		Payload: payload,
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

// tracks parses tracks from offer and adds additional info from State to them (like hints).
// Returns error if tracks are not allowed or not present in state. State should be
// submitted by peer before sending the offer. This is because WebRTC does not support passing
// additional info with offers.
func (p *Proxy) tracks(desc webrtc.SessionDescription) ([]event.Track, error) {
	getID := func(m sdp.Media) string {
		parts := strings.Split(m.Attribute("msid"), " ")
		if len(parts) != 2 {
			return ""
		}
		return parts[1]
	}
	isInactive := func(m sdp.Media) bool {
		return len(m.Attributes.Values("inactive")) != 0
	}

	msg, err := sdp.Decode([]byte(desc.SDP))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrValidation, err)
	}

	var tracks []event.Track
	for _, m := range msg.Medias {
		switch m.Description.Type {
		case mediaVideo, mediaAudio:
			if isInactive(m) {
				continue
			}
			t, ok := p.state.Track(getID(m))
			if !ok {
				return nil, fmt.Errorf("%w: track not in state", ErrValidation)
			}
			tracks = append(tracks, t)
		case mediaApplication:
		default:
			return nil, fmt.Errorf("%w: %s (%s)", ErrValidation, "media type not allowed", m.Description.Type)
		}
	}
	return tracks, nil
}

const (
	mediaVideo       = "video"
	mediaAudio       = "audio"
	mediaApplication = "application"
)

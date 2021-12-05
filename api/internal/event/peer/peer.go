package peer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aromancev/confa/internal/event"
	"github.com/aromancev/confa/internal/platform/grpcpool"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/proto/queue"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"gortc.io/sdp"
)

var (
	ErrValidation     = errors.New("validation error")
	ErrClosed         = errors.New("connection closed")
	ErrUnknownMessage = errors.New("unknown message")
)

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type MessageType string

const (
	TypeJoin     MessageType = "join"
	TypeAnswer   MessageType = "answer"
	TypeOffer    MessageType = "offer"
	TypeTrickle  MessageType = "trickle"
	TypeEvent    MessageType = "event"
	TypeEventAck MessageType = "event_ack"
	TypeState    MessageType = "state"
	TypeError    MessageType = "error"
)

type Message struct {
	RequestID string      `json:"requestId,omitempty"`
	Type      MessageType `json:"type"`
	Payload   interface{} `json:"payload"`
}

type EventAck struct {
	EventID string `json:"eventId"`
}

type State struct {
	Tracks map[string]event.Track `json:"tracks"`
}

func (s State) Validate() error {
	if len(s.Tracks) > 3 {
		return errors.New("no more than 3 tracks allowed")
	}
	for id, t := range s.Tracks {
		if id == "" {
			return fmt.Errorf("invalid track: id cannot be zero")
		}
		if err := t.Validate(); err != nil {
			return fmt.Errorf("invalid track: %w", err)
		}
	}
	return nil
}

func (m *Message) UnmarshalJSON(b []byte) error {
	var raw struct {
		RequestID string          `json:"requestId,omitempty"`
		T         MessageType     `json:"type"`
		P         json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	switch raw.T {
	case TypeEvent:
		var p event.Event
		if err := json.Unmarshal(raw.P, &p); err != nil {
			return err
		}
		m.Payload = p
	case TypeJoin:
		var p Join
		if err := json.Unmarshal(raw.P, &p); err != nil {
			return err
		}
		m.Payload = p
	case TypeOffer:
		var p Offer
		if err := json.Unmarshal(raw.P, &p); err != nil {
			return err
		}
		m.Payload = p
	case TypeAnswer:
		var p Answer
		if err := json.Unmarshal(raw.P, &p); err != nil {
			return err
		}
		m.Payload = p
	case TypeTrickle:
		var p Trickle
		if err := json.Unmarshal(raw.P, &p); err != nil {
			return err
		}
		m.Payload = p
	case TypeState:
		var p State
		if err := json.Unmarshal(raw.P, &p); err != nil {
			return err
		}
		m.Payload = p
	default:
		return ErrUnknownMessage
	}

	m.Type = raw.T
	m.RequestID = raw.RequestID
	return nil
}

type Signal interface {
	Join(context.Context, Join) error
	Trickle(context.Context, Trickle) error
	Offer(context.Context, Offer) error
	Answer(context.Context, Answer) error
	Receive(context.Context) (Message, error)
	Close(context.Context) error
}

type Peer struct {
	ctx                    context.Context
	cancel                 func()
	sfuPool                *grpcpool.Pool
	signal                 Signal
	out                    chan Message
	producer               Producer
	events                 event.Cursor
	userID, roomID         uuid.UUID
	state                  State
	eventsDone, signalDone chan struct{}
}

func NewPeer(ctx context.Context, userID, roomID uuid.UUID, sfuPool *grpcpool.Pool, events event.Cursor, producer Producer, maxMessages int) *Peer {
	ctx, cancel := context.WithCancel(ctx)
	p := &Peer{
		ctx:        ctx,
		cancel:     cancel,
		sfuPool:    sfuPool,
		out:        make(chan Message, maxMessages),
		eventsDone: make(chan struct{}),
		signalDone: make(chan struct{}),
		producer:   producer,
		events:     events,
		userID:     userID,
		roomID:     roomID,
	}
	err := p.emitEvent(ctx, uuid.New(), event.Payload{
		Type: event.TypePeerState,
		Payload: event.PayloadPeerState{
			Status: event.PeerJoined,
		},
	})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to emit peer status event.")
	}
	go p.pullEvents()
	return p
}

func (p *Peer) Send(ctx context.Context, msg Message) error {
	switch pl := msg.Payload.(type) {
	case event.Event:
		id := uuid.New()
		select {
		case p.out <- Message{
			RequestID: msg.RequestID,
			Type:      TypeEventAck,
			Payload:   EventAck{EventID: id.String()},
		}:
		default:
			log.Ctx(ctx).Info().Msg("Peer too slow. Evicting.")
			p.cancel()
			return fmt.Errorf("%w: %s", ErrValidation, "peer too slow")
		}
		return p.emitUserEvent(ctx, id, pl.Payload)
	case State:
		if err := pl.Validate(); err != nil {
			return fmt.Errorf("%w: %s", ErrValidation, err)
		}
		p.state = pl
		select {
		case p.out <- Message{
			RequestID: msg.RequestID,
			Type:      TypeState,
			Payload:   pl,
		}:
		default:
			log.Ctx(ctx).Info().Msg("Peer too slow. Evicting.")
			p.cancel()
			return fmt.Errorf("%w: %s", ErrValidation, "peer too slow")
		}
		return nil
	case Join:
		var err error
		p.signal, err = NewGRPCSignal(ctx, p.sfuPool)
		if err != nil {
			return err
		}
		go p.pullSignal()
		return p.signal.Join(ctx, pl)
	case Offer:
		if p.signal == nil {
			return fmt.Errorf("%w: must join before offer", ErrValidation)
		}
		tracks, err := p.tracks(pl.Description)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrValidation, err)
		}
		if err := p.signal.Offer(ctx, pl); err != nil {
			return err
		}
		return p.emitEvent(ctx, uuid.New(), event.Payload{
			Type: event.TypePeerState,
			Payload: event.PayloadPeerState{
				Tracks: tracks,
			},
		})
	case Answer:
		if p.signal == nil {
			return fmt.Errorf("%w: must join before answer", ErrValidation)
		}
		return p.signal.Answer(ctx, pl)
	case Trickle:
		if p.signal == nil {
			return fmt.Errorf("%w: must join before trickle", ErrValidation)
		}
		return p.signal.Trickle(ctx, pl)
	default:
		return ErrUnknownMessage
	}
}

func (p *Peer) Receive(ctx context.Context) (Message, error) {
	select {
	case msg, ok := <-p.out:
		if !ok {
			return Message{}, ErrClosed
		}
		return msg, nil
	case <-ctx.Done():
		return Message{}, ctx.Err()
	}
}

func (p *Peer) Close(ctx context.Context) error {
	err := p.emitEvent(ctx, uuid.New(), event.Payload{
		Type: event.TypePeerState,
		Payload: event.PayloadPeerState{
			Status: event.PeerLeft,
		},
	})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to emit peer status event.")
	}
	var eventsErr, signalErr error
	<-p.eventsDone
	eventsErr = p.events.Close(ctx)
	if p.signal != nil {
		<-p.signalDone
		signalErr = p.signal.Close(ctx)
	}
	if eventsErr != nil {
		return eventsErr
	}
	return signalErr
}

func (p *Peer) pullEvents() {
	defer close(p.eventsDone)
	defer p.cancel() // If pulling exits, something went wrong.

	ctx := p.ctx

	for {
		ev, err := p.events.Next(ctx)
		switch {
		case errors.Is(err, event.ErrUnknownEvent):
			log.Ctx(ctx).Debug().Msg("Skipping unknown room event.")
			continue
		case errors.Is(err, event.ErrCursorClosed), errors.Is(err, context.Canceled):
			log.Ctx(ctx).Debug().Msg("Event cursor was closed.")
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to pull an event from cursor.")
			return
		}

		select {
		case p.out <- Message{
			Type:    TypeEvent,
			Payload: ev,
		}:
		default:
			log.Ctx(ctx).Info().Msg("Peer too slow. Evicting.")
			return
		}
		log.Ctx(ctx).Debug().Str("messageType", string(TypeEvent)).Str("eventType", string(ev.Payload.Type)).Msg("RTC message pulled.")
	}
}

func (p *Peer) pullSignal() {
	defer close(p.signalDone)
	defer p.cancel() // If pulling exits, something went wrong.

	ctx := p.ctx

	for {
		msg, err := p.signal.Receive(ctx)
		switch {
		case errors.Is(err, ErrUnknownMessage):
			log.Ctx(p.ctx).Debug().Msg("Skipping unknown signal message.")
			continue
		case errors.Is(err, ErrClosed), errors.Is(err, context.Canceled):
			log.Ctx(p.ctx).Debug().Msg("Signal was closed.")
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to pull a message from signal.")
			return
		}

		select {
		case p.out <- msg:
		default:
			log.Ctx(ctx).Info().Msg("Peer too slow. Evicting.")
			return
		}
		log.Ctx(ctx).Debug().Str("messageType", string(msg.Type)).Msg("RTC message pulled.")
	}
}

func (p *Peer) emitUserEvent(ctx context.Context, id uuid.UUID, payload event.Payload) error {
	switch payload.Type {
	case event.TypeMessage:
	default:
		return fmt.Errorf("%w: invalid user-initiated event type (%s)", ErrValidation, payload.Type)
	}
	return p.emitEvent(ctx, id, payload)
}

func (p *Peer) emitEvent(ctx context.Context, id uuid.UUID, payload event.Payload) error {
	if err := payload.Validate(); err != nil {
		return fmt.Errorf("%w: %s", ErrValidation, err)
	}
	eventID, _ := id.MarshalBinary()
	roomID, _ := p.roomID.MarshalBinary()
	ownerID, _ := p.userID.MarshalBinary()
	buf, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	job := queue.EventJob{
		Id:      eventID,
		RoomId:  roomID,
		OwnerId: ownerID,
		Payload: buf,
	}
	body, err := queue.Marshal(&job, trace.ID(ctx))
	if err != nil {
		return err
	}
	_, err = p.producer.Put(ctx, queue.TubeEvent, body, beanstalk.PutParams{TTR: 10 * time.Second})
	return err
}

// tracks parses tracks from offer and adds additional info from State to them (like hints).
// Returns error if tracks are not allowed or not present in state. State should be
// submitted by peer before sending the offer. This is because WebRTC does not support passing
// additional info with offers.
func (p *Peer) tracks(desc webrtc.SessionDescription) (map[string]event.Track, error) {
	getID := func(m sdp.Media) string {
		parts := strings.Split(m.Attribute("msid"), " ")
		if len(parts) != 2 {
			return ""
		}
		return parts[1]
	}

	msg, err := sdp.Decode([]byte(desc.SDP))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrValidation, err)
	}

	tracks := make(map[string]event.Track)
	for _, m := range msg.Medias {
		switch m.Description.Type {
		case mediaVideo, mediaAudio:
			trackID := getID(m)
			t, ok := p.state.Tracks[trackID]
			if !ok {
				return nil, fmt.Errorf("%w: track not in state", ErrValidation)
			}
			tracks[trackID] = t
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

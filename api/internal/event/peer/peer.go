package peer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	TypeJoin    MessageType = "join"
	TypeAnswer  MessageType = "answer"
	TypeOffer   MessageType = "offer"
	TypeTrickle MessageType = "trickle"
	TypeEvent   MessageType = "event"
	TypeError   MessageType = "error"
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

func (m *Message) UnmarshalJSON(b []byte) error {
	var raw struct {
		T MessageType     `json:"type"`
		P json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	switch raw.T {
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
	default:
		return ErrUnknownMessage
	}

	m.Type = raw.T
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
	ctx            context.Context
	cancel         func()
	sfuPool        *grpcpool.Pool
	signal         Signal
	out            chan Message
	producer       Producer
	events         event.Cursor
	userID, roomID uuid.UUID
}

func NewPeer(ctx context.Context, userID, roomID uuid.UUID, sfuPool *grpcpool.Pool, events event.Cursor, producer Producer, maxMessages int) *Peer {
	ctx, cancel := context.WithCancel(ctx)
	p := &Peer{
		ctx:      ctx,
		cancel:   cancel,
		sfuPool:  sfuPool,
		out:      make(chan Message, maxMessages),
		producer: producer,
		events:   events,
		userID:   userID,
		roomID:   roomID,
	}
	if err := p.emitStatus(ctx, event.PeerJoined); err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to emit peer status event.")
	}
	go p.pullEvents()
	return p
}

func (p *Peer) Send(ctx context.Context, msg Message) error {
	switch pl := msg.Payload.(type) {
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
		if err := validateOffer(pl.Description); err != nil {
			return fmt.Errorf("%w: %s", ErrValidation, err)
		}
		return p.signal.Offer(ctx, pl)
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
	if err := p.emitStatus(ctx, event.PeerLeft); err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to emit peer status event.")
	}
	var eventsErr, signalErr error
	eventsErr = p.events.Close(ctx)
	if p.signal != nil {
		signalErr = p.signal.Close(ctx)
	}
	if eventsErr != nil {
		return eventsErr
	}
	return signalErr
}

func (p *Peer) pullEvents() {
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

func (p *Peer) emitStatus(ctx context.Context, status event.PeerStatus) error {
	eventID, _ := uuid.New().MarshalBinary()
	roomID, _ := p.roomID.MarshalBinary()
	ownerID, _ := p.userID.MarshalBinary()
	body, err := queue.Marshal(&queue.EventJob{
		Id:      eventID,
		RoomId:  roomID,
		OwnerId: ownerID,
		Event: &queue.EventJob_PeerStatus_{
			PeerStatus: &queue.EventJob_PeerStatus{
				Status: string(status),
			},
		},
	}, trace.ID(ctx))
	if err != nil {
		return err
	}
	_, err = p.producer.Put(ctx, queue.TubeEvent, body, beanstalk.PutParams{TTR: 10 * time.Second})
	return err
}

func validateOffer(desc webrtc.SessionDescription) error {
	msg, err := sdp.Decode([]byte(desc.SDP))
	if err != nil {
		return fmt.Errorf("%w %s", ErrValidation, err)
	}

	var video, audio, app uint
	for _, m := range msg.Medias {
		switch m.Description.Type {
		case mediaVideo:
			video++
		case mediaAudio:
			audio++
		case mediaApplication:
			app++
		default:
			return fmt.Errorf("%w: %s (%s)", ErrValidation, "media type not allowed", m.Description.Type)
		}
	}

	if video > 2 {
		return fmt.Errorf("%w %s", ErrValidation, "maximum 2 video tracks allowed")
	}
	if audio > 2 {
		return fmt.Errorf("%w %s", ErrValidation, "maximum 2 audio tracks allowed")
	}
	if app > 1 {
		return fmt.Errorf("%w %s", ErrValidation, "maximum 1 application track allowed")
	}
	return nil
}

const (
	mediaVideo       = "video"
	mediaAudio       = "audio"
	mediaApplication = "application"
)

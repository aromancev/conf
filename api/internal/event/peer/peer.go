package peer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aromancev/confa/internal/platform/grpcpool"
	"github.com/pion/webrtc/v3"
	"gortc.io/sdp"
)

var (
	ErrValidation     = errors.New("validation error")
	ErrClosed         = errors.New("connection closed")
	ErrUnknownMessage = errors.New("unknown message")
)

type MessageType string

const (
	TypeJoin    MessageType = "join"
	TypeAnswer  MessageType = "answer"
	TypeOffer   MessageType = "offer"
	TypeTrickle MessageType = "trickle"
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

type Peer struct {
	signal *Signal
}

func NewPeer(ctx context.Context, sfuPool *grpcpool.Pool) (*Peer, error) {
	signal, err := NewSignal(ctx, sfuPool)
	if err != nil {
		return nil, err
	}
	return &Peer{
		signal: signal,
	}, nil
}

func (p *Peer) Send(ctx context.Context, msg Message) error {
	switch pl := msg.Payload.(type) {
	case Join:
		return p.signal.Join(ctx, pl)
	case Offer:
		if err := validateOffer(pl.Description); err != nil {
			return fmt.Errorf("%w: %s", ErrValidation, err)
		}
		return p.signal.Offer(ctx, pl)
	case Answer:
		return p.signal.Answer(ctx, pl)
	case Trickle:
		return p.signal.Trickle(ctx, pl)
	default:
		return ErrUnknownMessage
	}
}

func (p *Peer) Receive(ctx context.Context) (Message, error) {
	msg, err := p.signal.Receive(ctx)
	switch {
	case errors.Is(err, ErrClosed):
		return Message{}, ErrClosed
	case errors.Is(err, ErrUnknownMessage):
		return Message{}, ErrUnknownMessage
	case err != nil:
		return Message{}, err
	}

	switch msg.(type) {
	case Answer:
		return Message{Type: TypeAnswer, Payload: msg}, nil
	case Offer:
		return Message{Type: TypeOffer, Payload: msg}, nil
	case Trickle:
		return Message{Type: TypeTrickle, Payload: msg}, nil
	}

	return Message{}, errors.New("unknown msg type")
}

func (p *Peer) Close(ctx context.Context) error {
	return p.signal.Close(ctx)
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

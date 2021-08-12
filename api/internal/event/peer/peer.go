package peer

import (
	"context"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v3"
	"gortc.io/sdp"
)

var (
	ErrValidation = errors.New("validation error")
)

type Signal interface {
	OnAnswer(func(webrtc.SessionDescription))
	OnOffer(func(webrtc.SessionDescription))
	OnTrickle(func(cand webrtc.ICECandidateInit, target int))

	Join(sid, uid string, offer webrtc.SessionDescription) error
	Offer(webrtc.SessionDescription)
	Trickle(cand webrtc.ICECandidateInit, target int)
	Answer(webrtc.SessionDescription)
}

type Peer struct {
	signal            Signal
	onAnswer, onOffer func(webrtc.SessionDescription)
	onTrickle         func(webrtc.ICECandidateInit, int)
}

func NewPeer(sig Signal) *Peer {
	peer := &Peer{
		signal: sig,
	}

	sig.OnOffer(func(desc webrtc.SessionDescription) {
		if peer.onOffer != nil {
			peer.onOffer(desc)
		}
	})
	sig.OnAnswer(func(desc webrtc.SessionDescription) {
		if peer.onAnswer != nil {
			peer.onAnswer(desc)
		}
	})
	sig.OnTrickle(func(cand webrtc.ICECandidateInit, target int) {
		if peer.onTrickle != nil {
			peer.onTrickle(cand, target)
		}
	})
	return peer
}

func (p *Peer) OnAnswer(f func(webrtc.SessionDescription)) {
	p.onAnswer = f
}

func (p *Peer) OnOffer(f func(webrtc.SessionDescription)) {
	p.onOffer = f
}

func (p *Peer) OnTrickle(f func(webrtc.ICECandidateInit, int)) {
	p.onTrickle = f
}

func (p *Peer) Join(_ context.Context, sid, uid string, offer webrtc.SessionDescription) error {
	return p.signal.Join(sid, uid, offer)
}

func (p *Peer) Offer(_ context.Context, desc webrtc.SessionDescription) error {
	err := validateOffer(desc)
	if err != nil {
		return err
	}

	p.signal.Offer(desc)
	return nil
}

func (p *Peer) Trickle(_ context.Context, cand webrtc.ICECandidateInit, target int) error {
	p.signal.Trickle(cand, target)
	return nil
}

func (p *Peer) Answer(_ context.Context, desc webrtc.SessionDescription) error {
	p.signal.Answer(desc)
	return nil
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

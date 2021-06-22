package rtc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v3"
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

type Session struct {
	signal            Signal
	onAnswer, onOffer func(webrtc.SessionDescription)
	onTrickle         func(webrtc.ICECandidateInit, int)
}

func NewSession(sig Signal) *Session {
	sess := &Session{
		signal: sig,
	}

	sig.OnOffer(func(desc webrtc.SessionDescription) {
		if sess.onOffer != nil {
			sess.onOffer(desc)
		}
	})
	sig.OnAnswer(func(desc webrtc.SessionDescription) {
		if sess.onAnswer != nil {
			sess.onAnswer(desc)
		}
	})
	sig.OnTrickle(func(cand webrtc.ICECandidateInit, target int) {
		if sess.onTrickle != nil {
			sess.onTrickle(cand, target)
		}
	})
	return sess
}

func (s *Session) OnAnswer(f func(webrtc.SessionDescription)) {
	s.onAnswer = f
}

func (s *Session) OnOffer(f func(webrtc.SessionDescription)) {
	s.onOffer = f
}

func (s *Session) OnTrickle(f func(webrtc.ICECandidateInit, int)) {
	s.onTrickle = f
}

func (s *Session) Join(_ context.Context, sid, uid string, offer webrtc.SessionDescription) error {
	return s.signal.Join(sid, uid, offer)
}

func (s *Session) Offer(_ context.Context, desc webrtc.SessionDescription) error {
	_, err := parseOffer(desc)
	if err != nil {
		return err
	}

	s.signal.Offer(desc)
	return nil
}

func (s *Session) Trickle(_ context.Context, cand webrtc.ICECandidateInit, target int) error {
	s.signal.Trickle(cand, target)
	return nil
}

func (s *Session) Answer(_ context.Context, desc webrtc.SessionDescription) error {
	s.signal.Answer(desc)
	return nil
}

type track struct {
	id string
}

type offer struct {
	videos, audios, apps []track
}

func parseOffer(desc webrtc.SessionDescription) (offer, error) {
	trackID := func(d *sdp.MediaDescription) string {
		attr, ok := d.Attribute("msid")
		if !ok {
			return ""
		}
		parts := strings.Split(attr, " ")
		if len(parts) != 2 {
			return ""
		}
		return parts[1]
	}

	var off offer
	parsed, err := desc.Unmarshal()
	if err != nil {
		return offer{}, err
	}
	for _, m := range parsed.MediaDescriptions {
		t := track{
			id: trackID(m),
		}
		switch m.MediaName.Media {
		case mediaAudio:
			if t.id == "" {
				return offer{}, errors.New("track id should not be zero")
			}
			off.audios = append(off.audios, t)
		case mediaVideo:
			if t.id == "" {
				return offer{}, errors.New("track id should not be zero")
			}
			off.videos = append(off.videos, t)
		case mediaApplication:
			off.apps = append(off.apps, t)
		default:
			return offer{}, fmt.Errorf("media not supported: %s", m.MediaName.Media)
		}
	}

	if len(off.videos) > 2 {
		return offer{}, errors.New("too many video tracks")
	}
	if len(off.audios) > 1 {
		return offer{}, errors.New("too many audio tracks")
	}
	if len(off.apps) > 1 {
		return offer{}, errors.New("too many application tracks")
	}

	return off, nil
}

const (
	mediaVideo       = "video"
	mediaAudio       = "audio"
	mediaApplication = "application"
)

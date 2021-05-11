package rtc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v3"
	grpcpool "github.com/processout/grpc-go-pool"

	"github.com/aromancev/confa/proto/avp"
)

var (
	ErrValidation = errors.New("validation error")
)

type SFU interface {
	Join(ctx context.Context, sid, uid string, offer webrtc.SessionDescription) (webrtc.SessionDescription, error)
	Offer(ctx context.Context, offer webrtc.SessionDescription) (webrtc.SessionDescription, error)
}

type Session struct {
	sfu  SFU
	pool *grpcpool.Pool

	sessionID string
}

func NewSession(pool *grpcpool.Pool, sfu SFU) *Session {
	return &Session{
		pool: pool,
		sfu:  sfu,
	}
}

func (s *Session) Join(ctx context.Context, sid, uid string, offer webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	s.sessionID = sid
	return s.sfu.Join(ctx, sid, uid, offer)
}

func (s *Session) Offer(ctx context.Context, desc webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	if s.sessionID == "" {
		return webrtc.SessionDescription{}, errors.New("must join before offer")
	}

	off, err := parseOffer(desc)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	conn, err := s.pool.Get(ctx)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}
	defer conn.Close()

	client := avp.NewAVPClient(conn)
	_, err = client.Signal(ctx, &avp.Request{
		SessionId: s.sessionID,
		TrackId:   off.videos[0].id,
		Process:   avp.Process_SAVE,
	})
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	return s.sfu.Offer(ctx, desc)
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

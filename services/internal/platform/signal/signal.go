package signal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/pion/ion/proto/rtc"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrClosed         = errors.New("connection closed")
	ErrUnknownMessage = errors.New("unknown message")
)

type Join struct {
	SessionID   string
	UserID      string
	Description webrtc.SessionDescription
}

type Answer struct {
	Description webrtc.SessionDescription
}

type Offer struct {
	Description webrtc.SessionDescription
}

type Trickle struct {
	Candidate webrtc.ICECandidateInit
	Target    int
}

type Message struct {
	Join    *Join
	Answer  *Answer
	Offer   *Offer
	Trickle *Trickle
}

type GRPCSignal struct {
	m      sync.Mutex
	stream rtc.RTC_SignalClient
}

func NewGRPCSignal(ctx context.Context, conn *grpc.ClientConn) (*GRPCSignal, error) {
	stream, err := rtc.NewRTCClient(conn).Signal(ctx)
	if err != nil {
		return nil, err
	}
	return &GRPCSignal{
		stream: stream,
	}, nil
}

func (s *GRPCSignal) Send(ctx context.Context, msg Message) error {
	s.m.Lock()
	defer s.m.Unlock()

	switch {
	case msg.Join != nil:
		pl := msg.Join
		return s.stream.Send(&rtc.Request{
			Payload: &rtc.Request_Join{
				Join: &rtc.JoinRequest{
					Sid: pl.SessionID,
					Uid: pl.UserID,
					Description: &rtc.SessionDescription{
						Target: rtc.Target_PUBLISHER,
						Type:   pl.Description.Type.String(),
						Sdp:    pl.Description.SDP,
					},
				},
			},
		})

	case msg.Offer != nil:
		pl := msg.Offer
		return s.stream.Send(&rtc.Request{
			Payload: &rtc.Request_Description{
				Description: &rtc.SessionDescription{
					Target: rtc.Target_PUBLISHER,
					Type:   pl.Description.Type.String(),
					Sdp:    pl.Description.SDP,
				},
			},
		})

	case msg.Answer != nil:
		pl := msg.Answer
		return s.stream.Send(&rtc.Request{
			Payload: &rtc.Request_Description{
				Description: &rtc.SessionDescription{
					Target: rtc.Target_SUBSCRIBER,
					Type:   pl.Description.Type.String(),
					Sdp:    pl.Description.SDP,
				},
			},
		})

	case msg.Trickle != nil:
		pl := msg.Trickle
		cand, err := json.Marshal(pl.Candidate)
		if err != nil {
			return fmt.Errorf("failed to marshal trickle: %w", err)
		}
		return s.stream.Send(&rtc.Request{
			Payload: &rtc.Request_Trickle{
				Trickle: &rtc.Trickle{
					Init:   string(cand),
					Target: rtc.Target(pl.Target),
				},
			},
		})
	}

	return ErrUnknownMessage
}

// Receive fetches an incoming message from SFU. It is not safe to call concurrently.
func (s *GRPCSignal) Receive(ctx context.Context) (Message, error) {
	res, err := s.stream.Recv()
	errStatus, _ := status.FromError(err)
	switch {
	case errors.Is(err, io.EOF), errStatus.Code() == codes.Canceled:
		return Message{}, ErrClosed
	case err != nil:
		return Message{}, err
	}

	switch payload := res.Payload.(type) {
	case *rtc.Reply_Join:
		if !payload.Join.Success {
			return Message{}, fmt.Errorf("failed to join: %s", payload.Join.Error.String())
		}
		return Message{
			Answer: &Answer{
				Description: webrtc.SessionDescription{
					Type: webrtc.NewSDPType(payload.Join.Description.Type),
					SDP:  payload.Join.Description.Sdp,
				},
			},
		}, nil

	case *rtc.Reply_Description:
		desc := webrtc.SessionDescription{
			Type: webrtc.NewSDPType(payload.Description.Type),
			SDP:  payload.Description.Sdp,
		}
		switch desc.Type {
		case webrtc.SDPTypeOffer:
			return Message{
				Offer: &Offer{
					Description: desc,
				},
			}, nil
		case webrtc.SDPTypeAnswer:
			return Message{
				Answer: &Answer{
					Description: desc,
				},
			}, nil
		}

	case *rtc.Reply_Trickle:
		var c webrtc.ICECandidateInit
		_ = json.Unmarshal([]byte(payload.Trickle.Init), &c) // Init unmarshal errors are ont critical.
		return Message{
			Trickle: &Trickle{
				Candidate: c,
				Target:    int(payload.Trickle.Target),
			},
		}, nil
	}

	return Message{}, ErrUnknownMessage
}

func (s *GRPCSignal) Close(_ context.Context) error {
	s.m.Lock()
	defer s.m.Unlock()
	return s.stream.CloseSend()
}

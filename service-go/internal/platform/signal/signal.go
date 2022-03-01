package signal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	pb "github.com/pion/ion-sfu/cmd/signal/grpc/proto"
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
	stream pb.SFU_SignalClient
}

func NewGRPCSignal(ctx context.Context, conn *grpc.ClientConn) (*GRPCSignal, error) {
	stream, err := pb.NewSFUClient(conn).Signal(ctx)
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
		desc, err := json.Marshal(pl.Description)
		if err != nil {
			return fmt.Errorf("failed to marshal join: %w", err)
		}
		return s.stream.Send(&pb.SignalRequest{
			Payload: &pb.SignalRequest_Join{
				Join: &pb.JoinRequest{
					Sid:         pl.SessionID,
					Uid:         pl.UserID,
					Description: desc,
				},
			},
		})
	case msg.Offer != nil:
		pl := msg.Offer
		desc, err := json.Marshal(pl.Description)
		if err != nil {
			return fmt.Errorf("failed to marshal offer: %w", err)
		}
		return s.stream.Send(&pb.SignalRequest{
			Payload: &pb.SignalRequest_Description{
				Description: desc,
			},
		})
	case msg.Answer != nil:
		pl := msg.Answer
		desc, err := json.Marshal(pl.Description)
		if err != nil {
			return fmt.Errorf("failed to marshal answer: %w", err)
		}
		return s.stream.Send(&pb.SignalRequest{
			Payload: &pb.SignalRequest_Description{
				Description: desc,
			},
		})
	case msg.Trickle != nil:
		pl := msg.Trickle
		cand, err := json.Marshal(pl.Candidate)
		if err != nil {
			return fmt.Errorf("failed to marshal trickle: %w", err)
		}
		return s.stream.Send(&pb.SignalRequest{
			Payload: &pb.SignalRequest_Trickle{
				Trickle: &pb.Trickle{
					Init:   string(cand),
					Target: pb.Trickle_Target(pl.Target),
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
	case *pb.SignalReply_Join:
		var s webrtc.SessionDescription
		err := json.Unmarshal(payload.Join.Description, &s)
		if err != nil {
			return Message{}, fmt.Errorf("failed to unmarshal session: %w", err)
		}
		return Message{Answer: &Answer{Description: s}}, nil

	case *pb.SignalReply_Description:
		var s webrtc.SessionDescription
		err := json.Unmarshal(payload.Description, &s)
		if err != nil {
			return Message{}, fmt.Errorf("failed to unmarshal session: %w", err)
		}
		switch s.Type {
		case webrtc.SDPTypeOffer:
			return Message{Offer: &Offer{Description: s}}, nil
		case webrtc.SDPTypeAnswer:
			return Message{Answer: &Answer{Description: s}}, nil
		}

	case *pb.SignalReply_Trickle:
		var c webrtc.ICECandidateInit
		_ = json.Unmarshal([]byte(payload.Trickle.Init), &c) // Init unmarshal errors are ont critical.
		return Message{Trickle: &Trickle{Candidate: c, Target: int(payload.Trickle.Target)}}, nil
	}

	return Message{}, ErrUnknownMessage
}

func (s *GRPCSignal) Close(_ context.Context) error {
	s.m.Lock()
	defer s.m.Unlock()
	return s.stream.CloseSend()
}

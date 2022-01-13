package signal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/aromancev/confa/internal/event/peer"
	pb "github.com/pion/ion-sfu/cmd/signal/grpc/proto"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCSignal struct {
	m      sync.Mutex
	stream pb.SFU_SignalClient
	conn   *grpc.ClientConn
}

func NewGRPCSignal(ctx context.Context, conn *grpc.ClientConn) *GRPCSignal {
	return &GRPCSignal{
		conn: conn,
	}
}

func (s *GRPCSignal) Connect(ctx context.Context) error {
	var err error
	s.stream, err = pb.NewSFUClient(s.conn).Signal(ctx)
	return err
}

func (s *GRPCSignal) Join(_ context.Context, req peer.Join) error {
	desc, err := json.Marshal(req.Description)
	if err != nil {
		return fmt.Errorf("failed to marshal join: %w", err)
	}

	s.m.Lock()
	defer s.m.Unlock()
	if s.stream == nil {
		return peer.ErrClosed
	}
	return s.stream.Send(
		&pb.SignalRequest{
			Payload: &pb.SignalRequest_Join{
				Join: &pb.JoinRequest{
					Sid:         req.SessionID,
					Uid:         req.UserID,
					Description: desc,
				},
			},
		},
	)
}

func (s *GRPCSignal) Trickle(_ context.Context, req peer.Trickle) error {
	bytes, err := json.Marshal(req.Candidate)
	if err != nil {
		return fmt.Errorf("failed to marshal trickle: %w", err)
	}

	s.m.Lock()
	defer s.m.Unlock()
	if s.stream == nil {
		return peer.ErrClosed
	}
	return s.stream.Send(&pb.SignalRequest{
		Payload: &pb.SignalRequest_Trickle{
			Trickle: &pb.Trickle{
				Init:   string(bytes),
				Target: pb.Trickle_Target(req.Target),
			},
		},
	})
}

func (s *GRPCSignal) Offer(_ context.Context, req peer.Offer) error {
	desc, err := json.Marshal(req.Description)
	if err != nil {
		return fmt.Errorf("failed to marshal offer: %w", err)
	}

	s.m.Lock()
	defer s.m.Unlock()
	if s.stream == nil {
		return peer.ErrClosed
	}
	return s.stream.Send(
		&pb.SignalRequest{
			Payload: &pb.SignalRequest_Description{
				Description: desc,
			},
		},
	)
}

func (s *GRPCSignal) Answer(_ context.Context, req peer.Answer) error {
	desc, err := json.Marshal(req.Description)
	if err != nil {
		return fmt.Errorf("failed to marshal answer: %w", err)
	}

	s.m.Lock()
	defer s.m.Unlock()
	if s.stream == nil {
		return peer.ErrClosed
	}
	return s.stream.Send(
		&pb.SignalRequest{
			Payload: &pb.SignalRequest_Description{
				Description: desc,
			},
		},
	)
}

// Receive fetches an incoming message from SFU. It is not safe to call concurrently.
func (s *GRPCSignal) Receive(ctx context.Context) (peer.Message, error) {
	if s.stream == nil {
		return peer.Message{}, peer.ErrClosed
	}

	res, err := s.stream.Recv()
	errStatus, _ := status.FromError(err)
	switch {
	case errors.Is(err, io.EOF), errStatus.Code() == codes.Canceled:
		return peer.Message{}, peer.ErrClosed
	case err != nil:
		return peer.Message{}, err
	}

	switch payload := res.Payload.(type) {
	case *pb.SignalReply_Join:
		var s webrtc.SessionDescription
		err := json.Unmarshal(payload.Join.Description, &s)
		if err != nil {
			return peer.Message{}, fmt.Errorf("failed to unmarshal session: %w", err)
		}
		return peer.Message{
			Type:    peer.TypeAnswer,
			Payload: peer.Answer{Description: s},
		}, nil

	case *pb.SignalReply_Description:
		var s webrtc.SessionDescription
		err := json.Unmarshal(payload.Description, &s)
		if err != nil {
			return peer.Message{}, fmt.Errorf("failed to unmarshal session: %w", err)
		}
		switch s.Type {
		case webrtc.SDPTypeOffer:
			return peer.Message{
				Type:    peer.TypeOffer,
				Payload: peer.Offer{Description: s},
			}, nil
		case webrtc.SDPTypeAnswer:
			return peer.Message{
				Type:    peer.TypeAnswer,
				Payload: peer.Answer{Description: s},
			}, nil
		}

	case *pb.SignalReply_Trickle:
		var c webrtc.ICECandidateInit
		_ = json.Unmarshal([]byte(payload.Trickle.Init), &c) // Init unmarshal errors are ont critical.
		return peer.Message{
			Type:    peer.TypeTrickle,
			Payload: peer.Trickle{Candidate: c, Target: int(payload.Trickle.Target)},
		}, nil
	}

	return peer.Message{}, peer.ErrUnknownMessage
}

func (s *GRPCSignal) Close(_ context.Context) error {
	s.m.Lock()
	defer s.m.Unlock()
	return s.stream.CloseSend()
}

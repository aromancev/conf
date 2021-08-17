package peer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/aromancev/confa/internal/platform/grpcpool"
	pb "github.com/pion/ion-sfu/cmd/signal/grpc/proto"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Join struct {
	SessionID   string                    `json:"sessionId"`
	UserID      string                    `json:"userId"`
	Description webrtc.SessionDescription `json:"description"`
}

type Answer struct {
	Description webrtc.SessionDescription `json:"description"`
}

type Offer struct {
	Description webrtc.SessionDescription `json:"description"`
}

type Trickle struct {
	Candidate webrtc.ICECandidateInit `json:"candidate"`
	Target    int                     `json:"target"`
}

type GRPCSignal struct {
	m      sync.Mutex
	conn   *grpcpool.ClientConn
	client pb.SFUClient
	stream pb.SFU_SignalClient
}

func NewGRPCSignal(ctx context.Context, sfuPool *grpcpool.Pool) (*GRPCSignal, error) {
	conn, err := sfuPool.Get(ctx)
	if err != nil {
		return nil, err
	}
	client := pb.NewSFUClient(conn)
	stream, err := client.Signal(ctx)
	if err != nil {
		return nil, err
	}
	return &GRPCSignal{
		conn:   conn,
		client: client,
		stream: stream,
	}, nil
}

func (s *GRPCSignal) Join(_ context.Context, req Join) error {
	desc, err := json.Marshal(req.Description)
	if err != nil {
		return fmt.Errorf("failed to marshal join: %w", err)
	}

	s.m.Lock()
	defer s.m.Unlock()
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

func (s *GRPCSignal) Trickle(_ context.Context, req Trickle) error {
	bytes, err := json.Marshal(req.Candidate)
	if err != nil {
		return fmt.Errorf("failed to marshal trickle: %w", err)
	}

	s.m.Lock()
	defer s.m.Unlock()
	return s.stream.Send(&pb.SignalRequest{
		Payload: &pb.SignalRequest_Trickle{
			Trickle: &pb.Trickle{
				Init:   string(bytes),
				Target: pb.Trickle_Target(req.Target),
			},
		},
	})
}

func (s *GRPCSignal) Offer(_ context.Context, req Offer) error {
	desc, err := json.Marshal(req.Description)
	if err != nil {
		return fmt.Errorf("failed to marshal offer: %w", err)
	}

	s.m.Lock()
	defer s.m.Unlock()
	return s.stream.Send(
		&pb.SignalRequest{
			Payload: &pb.SignalRequest_Description{
				Description: desc,
			},
		},
	)
}

func (s *GRPCSignal) Answer(_ context.Context, req Answer) error {
	desc, err := json.Marshal(req.Description)
	if err != nil {
		return fmt.Errorf("failed to marshal answer: %w", err)
	}

	s.m.Lock()
	defer s.m.Unlock()
	return s.stream.Send(
		&pb.SignalRequest{
			Payload: &pb.SignalRequest_Description{
				Description: desc,
			},
		},
	)
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
		return Message{
			Type:    TypeAnswer,
			Payload: Answer{Description: s},
		}, nil

	case *pb.SignalReply_Description:
		var s webrtc.SessionDescription
		err := json.Unmarshal(payload.Description, &s)
		if err != nil {
			return Message{}, fmt.Errorf("failed to unmarshal session: %w", err)
		}
		switch s.Type {
		case webrtc.SDPTypeOffer:
			return Message{
				Type:    TypeOffer,
				Payload: Offer{Description: s},
			}, nil
		case webrtc.SDPTypeAnswer:
			return Message{
				Type:    TypeAnswer,
				Payload: Answer{Description: s},
			}, nil
		}

	case *pb.SignalReply_Trickle:
		var c webrtc.ICECandidateInit
		_ = json.Unmarshal([]byte(payload.Trickle.Init), &c) // Init unmarshal errors are ont critical.
		return Message{
			Type:    TypeTrickle,
			Payload: Trickle{Candidate: c, Target: int(payload.Trickle.Target)},
		}, nil
	}

	return Message{}, ErrUnknownMessage
}

func (s *GRPCSignal) Close(_ context.Context) error {
	s.m.Lock()
	defer s.m.Unlock()
	_ = s.stream.CloseSend()
	return s.conn.Close()
}

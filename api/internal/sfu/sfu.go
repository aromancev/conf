package sfu

import (
	"context"
	"fmt"
	"sync"

	"github.com/pion/webrtc/v3"
	grpcpool "github.com/processout/grpc-go-pool"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aromancev/confa/proto/sfu"
)

type Config struct {
	OnOffer   func(offer webrtc.SessionDescription)
	OnTrickle func(target int32, candidate webrtc.ICECandidateInit)
}

type Peer struct {
	lock        sync.RWMutex
	conn        *grpcpool.ClientConn
	stream      sfu.SFU_SignalClient
	requestID   uint32
	join, offer request

	onOffer   func(offer webrtc.SessionDescription)
	onTrickle func(target int32, candidate webrtc.ICECandidateInit)
}

func NewPeer(ctx context.Context, pool *grpcpool.Pool, cfg Config) (*Peer, error) {
	conn, err := pool.Get(ctx)
	if err != nil {
		return nil, err
	}
	client := sfu.NewSFUClient(conn)
	stream, err := client.Signal(ctx)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open GRPC stream: %w", err)
	}

	peer := &Peer{
		conn:      conn,
		stream:    stream,
		onOffer:   cfg.OnOffer,
		onTrickle: cfg.OnTrickle,
	}

	go peer.serve()
	return peer, nil
}

func (s *Peer) Join(ctx context.Context, sid string, offer webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	s.lock.Lock()
	s.requestID++
	id := s.requestID
	req := s.join.reset(id)
	err := s.stream.Send(&sfu.SignalRequest{
		Id: id,
		Payload: &sfu.SignalRequest_Join{
			Join: &sfu.SignalRequest_JoinSession{
				SessionId:   sid,
				Description: sfu.SessionDescriptionFromRTC(offer),
			},
		},
	})
	s.lock.Unlock()
	if err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("failed to send mesage to SFU")
	}

	var reply *sfu.SignalReply
	var ok bool
	select {
	case reply, ok = <-req:
		if !ok {
			return webrtc.SessionDescription{}, fmt.Errorf("request to SFU canceled (id=%d)", id)
		}
	case <-ctx.Done():
		return webrtc.SessionDescription{}, ctx.Err()
	}

	payload, ok := reply.Payload.(*sfu.SignalReply_Join)
	if !ok {
		return webrtc.SessionDescription{}, fmt.Errorf("unexpected reply from SFU (id=%d)", id)
	}
	return webrtc.SessionDescription{
		Type: webrtc.SDPType(payload.Join.Type),
		SDP:  payload.Join.Sdp,
	}, nil
}

func (s *Peer) Offer(ctx context.Context, offer webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	s.lock.Lock()
	s.requestID++
	id := s.requestID
	req := s.offer.reset(id)
	err := s.stream.Send(&sfu.SignalRequest{
		Id: id,
		Payload: &sfu.SignalRequest_Offer{
			Offer: sfu.SessionDescriptionFromRTC(offer),
		},
	})
	s.lock.Unlock()
	if err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("failed to send mesage to SFU")
	}

	var reply *sfu.SignalReply
	var ok bool
	select {
	case reply, ok = <-req:
		if !ok {
			return webrtc.SessionDescription{}, fmt.Errorf("request to SFU canceled (id=%d)", id)
		}
	case <-ctx.Done():
		return webrtc.SessionDescription{}, ctx.Err()
	}
	payload, ok := reply.Payload.(*sfu.SignalReply_Offer)
	if !ok {
		return webrtc.SessionDescription{}, fmt.Errorf("unexpected reply from SFU (id=%d)", id)
	}
	return webrtc.SessionDescription{
		Type: webrtc.SDPType(payload.Offer.Type),
		SDP:  payload.Offer.Sdp,
	}, nil
}

func (s *Peer) Answer(_ context.Context, answer webrtc.SessionDescription) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.requestID++
	return s.stream.Send(&sfu.SignalRequest{
		Id: s.requestID,
		Payload: &sfu.SignalRequest_Answer{
			Answer: sfu.SessionDescriptionFromRTC(answer),
		},
	})
}

func (s *Peer) Trickle(_ context.Context, target int32, candidate webrtc.ICECandidateInit) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.requestID++
	return s.stream.Send(&sfu.SignalRequest{
		Id: s.requestID,
		Payload: &sfu.SignalRequest_Trickle{
			Trickle: &sfu.Trickle{
				Target:    target,
				Candidate: sfu.CandidateInitFromRTC(candidate),
			},
		},
	})
}

func (s *Peer) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	_ = s.stream.CloseSend()
	return s.conn.Close()
}

func (s *Peer) serve() {
	handle := func(reply *sfu.SignalReply) {
		switch payload := reply.Payload.(type) {
		case *sfu.SignalReply_Join:
			if reply.Id == 0 {
				log.Error().Msg("Unexpected message from SFU")
				return
			}
			s.lock.Lock()
			defer s.lock.Unlock()
			if s.join.id != reply.Id {
				log.Error().Msg("Unexpected reply from SFU")
				return
			}
			select {
			case s.join.resp <- reply:
			default:
				log.Error().Msg("Unwanted reply from SFU")
				return
			}

		case *sfu.SignalReply_Offer:
			if reply.Id == 0 {
				if s.onOffer != nil {
					s.onOffer(webrtc.SessionDescription{
						Type: webrtc.SDPType(payload.Offer.Type),
						SDP:  payload.Offer.Sdp,
					})
				}
				return
			}
			s.lock.Lock()
			defer s.lock.Unlock()
			if s.offer.id != reply.Id {
				log.Error().Msg("Unexpected reply from SFU")
				return
			}
			select {
			case s.offer.resp <- reply:
			default:
				log.Error().Msg("Unwanted reply from SFU")
				return
			}

		case *sfu.SignalReply_Trickle:
			if s.onTrickle != nil {
				s.onTrickle(payload.Trickle.Target, sfu.CandidateInitToRTC(payload.Trickle.Candidate))
			}
		}
	}

	for {
		reply, err := s.stream.Recv()
		if err != nil {
			errStatus, _ := status.FromError(err)
			if errStatus.Code() == codes.Canceled {
				break
			}
			log.Err(err).Msg("Failed to receive from stream.")
			return
		}

		handle(reply)
	}
	log.Info().Msg("Peer disconnected.")
}

type request struct {
	id   uint32
	resp chan *sfu.SignalReply
}

func (r *request) reset(id uint32) chan *sfu.SignalReply {
	r.id = id
	if r.resp != nil {
		close(r.resp)
	}
	resp := make(chan *sfu.SignalReply, 1)
	r.resp = resp
	return resp
}

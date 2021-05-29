package sfu

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/pion/webrtc/v3"
	grpcpool "github.com/processout/grpc-go-pool"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Peer struct {
	lock        sync.RWMutex
	conn        *grpcpool.ClientConn
	stream      SFU_SignalClient
	requestID   uint32
	join, offer request

	onOffer   func(offer webrtc.SessionDescription)
	onTrickle func(target int, candidate webrtc.ICECandidateInit)
}

func NewPeer(ctx context.Context, pool *grpcpool.Pool) (*Peer, error) {
	conn, err := pool.Get(ctx)
	if err != nil {
		return nil, err
	}
	client := NewSFUClient(conn)
	stream, err := client.Signal(ctx)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open GRPC stream: %w", err)
	}

	peer := &Peer{
		conn:   conn,
		stream: stream,
	}

	go peer.serve()
	return peer, nil
}

func (p *Peer) OnOffer(f func(offer webrtc.SessionDescription)) {
	p.onOffer = f
}

func (p *Peer) OnTrickle(f func(target int, candidate webrtc.ICECandidateInit)) {
	p.onTrickle = f
}

func (p *Peer) Join(ctx context.Context, sid, uid string, offer webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	p.lock.Lock()
	p.requestID++
	id := p.requestID
	req := p.join.reset(id)
	err := p.stream.Send(&SignalRequest{
		Id: id,
		Payload: &SignalRequest_Join{
			Join: &SignalRequest_JoinSession{
				SessionId:   sid,
				UserId:      uid,
				Description: SessionDescriptionFromRTC(offer),
			},
		},
	})
	p.lock.Unlock()
	if err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("failed to send mesage to SFU: %w", err)
	}

	var reply *SignalReply
	var ok bool
	select {
	case reply, ok = <-req:
		if !ok {
			return webrtc.SessionDescription{}, fmt.Errorf("request to SFU canceled (id=%d)", id)
		}
	case <-ctx.Done():
		return webrtc.SessionDescription{}, ctx.Err()
	}

	payload, ok := reply.Payload.(*SignalReply_Join)
	if !ok {
		return webrtc.SessionDescription{}, fmt.Errorf("unexpected reply from SFU (id=%d)", id)
	}
	return webrtc.SessionDescription{
		Type: webrtc.SDPType(payload.Join.Type),
		SDP:  payload.Join.Sdp,
	}, nil
}

func (p *Peer) Offer(ctx context.Context, offer webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	p.lock.Lock()
	p.requestID++
	id := p.requestID
	req := p.offer.reset(id)
	err := p.stream.Send(&SignalRequest{
		Id: id,
		Payload: &SignalRequest_Offer{
			Offer: SessionDescriptionFromRTC(offer),
		},
	})
	p.lock.Unlock()
	if err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("failed to send mesage to SFU: %w", err)
	}

	var reply *SignalReply
	var ok bool
	select {
	case reply, ok = <-req:
		if !ok {
			return webrtc.SessionDescription{}, fmt.Errorf("request to SFU canceled (id=%d)", id)
		}
	case <-ctx.Done():
		return webrtc.SessionDescription{}, ctx.Err()
	}
	payload, ok := reply.Payload.(*SignalReply_Offer)
	if !ok {
		return webrtc.SessionDescription{}, fmt.Errorf("unexpected reply from SFU (id=%d)", id)
	}
	return webrtc.SessionDescription{
		Type: webrtc.SDPType(payload.Offer.Type),
		SDP:  payload.Offer.Sdp,
	}, nil
}

func (p *Peer) Answer(_ context.Context, answer webrtc.SessionDescription) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.requestID++
	return p.stream.Send(&SignalRequest{
		Id: p.requestID,
		Payload: &SignalRequest_Answer{
			Answer: SessionDescriptionFromRTC(answer),
		},
	})
}

func (p *Peer) Trickle(_ context.Context, target int, candidate webrtc.ICECandidateInit) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.requestID++
	return p.stream.Send(&SignalRequest{
		Id: p.requestID,
		Payload: &SignalRequest_Trickle{
			Trickle: &Trickle{
				Target:    int32(target),
				Candidate: CandidateInitFromRTC(candidate),
			},
		},
	})
}

func (p *Peer) Close() error {
	p.lock.Lock()
	defer p.lock.Unlock()

	_ = p.stream.CloseSend()
	return p.conn.Close()
}

func (p *Peer) serve() {
	defer func() {
		_ = p.Close()
	}()

	handle := func(reply *SignalReply) {
		switch payload := reply.Payload.(type) {
		case *SignalReply_Join:
			if reply.Id == 0 {
				log.Error().Msg("Unexpected message from SFU")
				return
			}
			p.lock.Lock()
			defer p.lock.Unlock()
			if p.join.id != reply.Id {
				log.Error().Msg("Unexpected reply from SFU")
				return
			}
			select {
			case p.join.resp <- reply:
			default:
				log.Error().Msg("Unwanted reply from SFU")
				return
			}

		case *SignalReply_Offer:
			if reply.Id == 0 {
				if p.onOffer != nil {
					p.onOffer(webrtc.SessionDescription{
						Type: webrtc.SDPType(payload.Offer.Type),
						SDP:  payload.Offer.Sdp,
					})
				}
				return
			}
			p.lock.Lock()
			defer p.lock.Unlock()
			if p.offer.id != reply.Id {
				log.Error().Msg("Unexpected reply from SFU")
				return
			}
			select {
			case p.offer.resp <- reply:
			default:
				log.Error().Msg("Unwanted reply from SFU")
				return
			}

		case *SignalReply_Trickle:
			if p.onTrickle != nil {
				p.onTrickle(int(payload.Trickle.Target), CandidateInitToRTC(payload.Trickle.Candidate))
			}
		}
	}

	for {
		reply, err := p.stream.Recv()
		errStatus, _ := status.FromError(err)
		switch {
		case errors.Is(err, io.EOF):
			return
		case errStatus.Code() == codes.Canceled:
			return
		case err != nil:
			log.Err(err).Msg("Failed to receive from stream.")
			return
		}
		handle(reply)
	}
}

type request struct {
	id   uint32
	resp chan *SignalReply
}

func (r *request) reset(id uint32) chan *SignalReply {
	r.id = id
	if r.resp != nil {
		close(r.resp)
	}
	resp := make(chan *SignalReply, 1)
	r.resp = resp
	return resp
}

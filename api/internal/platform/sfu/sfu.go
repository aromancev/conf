// This file is copied from github.com/pion/ion-sdk-go/signal.go with minor tweaks.
package sfu

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"

	pb "github.com/pion/ion-sfu/cmd/signal/grpc/proto"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Signal is a wrapper of grpc
type Signal struct {
	client pb.SFUClient
	stream pb.SFU_SignalClient

	onTrickle         func(webrtc.ICECandidateInit, int)
	onAnswer, onOffer func(webrtc.SessionDescription)

	sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	handleOnce sync.Once
}

// NewSignal create a grpc signaler
func NewSignal(ctx context.Context, sfuAddr string) (*Signal, error) {
	// Set up a connection to the sfu server.
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(connectCtx, sfuAddr, grpc.WithInsecure())
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to connect to SFU.")
		return nil, err
	}
	log.Ctx(ctx).Info().Msg("Connected to SFU")

	var s Signal
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.client = pb.NewSFUClient(conn)
	s.stream, err = s.client.Signal(s.ctx)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to establish signal connection.")
		return nil, err
	}
	return &s, nil
}

func (s *Signal) onSignalHandleOnce() {
	// onSignalHandle is wrapped in a once and only started after another public
	// method is called to ensure the user has the opportunity to register handlers
	s.handleOnce.Do(func() {
		err := s.onSignalHandle()
		if err != nil {
			log.Ctx(s.ctx).Err(err).Msg("Failed to handle signal.")
		}
	})
}

func (s *Signal) onSignalHandle() error {
	ctx := s.ctx

	for {
		res, err := s.stream.Recv()
		errStatus, _ := status.FromError(err)
		switch {
		case errors.Is(err, io.EOF), errStatus.Code() == codes.Canceled:
			log.Ctx(ctx).Info().Msg("WebRTC transport closed.")
			if err := s.stream.CloseSend(); err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to close SFU stream.")
				return err
			}
			return nil
		case err != nil:
			return err
		}

		switch payload := res.Payload.(type) {
		case *pb.SignalReply_Join:
			var sdp webrtc.SessionDescription
			err := json.Unmarshal(payload.Join.Description, &sdp)
			if err != nil {
				return err
			}

			if s.onAnswer != nil {
				s.onAnswer(sdp)
			}

		case *pb.SignalReply_Description:
			var sdp webrtc.SessionDescription
			err := json.Unmarshal(payload.Description, &sdp)
			if err != nil {
				return err
			}
			switch sdp.Type {
			case webrtc.SDPTypeOffer:
				if s.onOffer != nil {
					s.onOffer(sdp)
				}
			case webrtc.SDPTypeAnswer:
				if s.onAnswer != nil {
					s.onAnswer(sdp)
				}
			}

		case *pb.SignalReply_Trickle:
			var candidate webrtc.ICECandidateInit
			_ = json.Unmarshal([]byte(payload.Trickle.Init), &candidate)
			if s.onTrickle != nil {
				s.onTrickle(candidate, int(payload.Trickle.Target))
			}
		}
	}
}

func (s *Signal) OnAnswer(f func(webrtc.SessionDescription)) {
	s.onAnswer = f
}

func (s *Signal) OnOffer(f func(webrtc.SessionDescription)) {
	s.onOffer = f
}

func (s *Signal) OnTrickle(f func(webrtc.ICECandidateInit, int)) {
	s.onTrickle = f
}

func (s *Signal) Join(sid string, uid string, offer webrtc.SessionDescription) error {
	marshalled, err := json.Marshal(offer)
	if err != nil {
		return err
	}
	go s.onSignalHandleOnce()
	s.Lock()
	err = s.stream.Send(
		&pb.SignalRequest{
			Payload: &pb.SignalRequest_Join{
				Join: &pb.JoinRequest{
					Sid:         sid,
					Uid:         uid,
					Description: marshalled,
				},
			},
		},
	)
	s.Unlock()
	return err
}

func (s *Signal) Trickle(candidate webrtc.ICECandidateInit, target int) {
	ctx := s.ctx

	bytes, err := json.Marshal(candidate)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to marshal candidate.")
		return
	}
	go s.onSignalHandleOnce()
	s.Lock()
	err = s.stream.Send(&pb.SignalRequest{
		Payload: &pb.SignalRequest_Trickle{
			Trickle: &pb.Trickle{
				Init:   string(bytes),
				Target: pb.Trickle_Target(target),
			},
		},
	})
	s.Unlock()
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to send trickle to SFU.")
	}
}

func (s *Signal) Offer(sdp webrtc.SessionDescription) {
	ctx := s.ctx

	marshalled, err := json.Marshal(sdp)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to marshal session.")
		return
	}
	go s.onSignalHandleOnce()
	s.Lock()
	err = s.stream.Send(
		&pb.SignalRequest{
			Payload: &pb.SignalRequest_Description{
				Description: marshalled,
			},
		},
	)
	s.Unlock()
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to send offer to SFU.")
	}
}

func (s *Signal) Answer(sdp webrtc.SessionDescription) {
	ctx := s.ctx

	marshalled, err := json.Marshal(sdp)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to marshal session.")
		return
	}
	s.Lock()
	err = s.stream.Send(
		&pb.SignalRequest{
			Payload: &pb.SignalRequest_Description{
				Description: marshalled,
			},
		},
	)
	s.Unlock()
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to send answer to SFU.")
	}
}

func (s *Signal) Close() {
	s.cancel()
	s.onSignalHandleOnce()
	log.Ctx(s.ctx).Info().Msg("Signal closed.")
}

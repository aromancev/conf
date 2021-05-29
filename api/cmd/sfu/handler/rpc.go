package handler

import (
	"fmt"
	"io"
	"sync"

	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	proto "github.com/aromancev/confa/proto/sfu"
)

type SFU struct {
	proto.UnimplementedSFUServer
	SFU *sfu.SFU
}

func NewSFU(s *sfu.SFU) *SFU {
	return &SFU{SFU: s}
}

func (s *SFU) Signal(signal proto.SFU_SignalServer) error {
	stream := newStream(signal)

	peer := sfu.NewPeer(s.SFU)
	for {
		request, err := stream.Receive()

		if err != nil {
			_ = peer.Close()

			if err == io.EOF {
				return nil
			}

			errStatus, _ := status.FromError(err)
			if errStatus.Code() == codes.Canceled {
				return nil
			}

			return err
		}

		switch payload := request.Payload.(type) {
		case *proto.SignalRequest_Join:
			peer.OnOffer = func(desc *webrtc.SessionDescription) {
				err = stream.Notify(&proto.SignalReply{
					Payload: &proto.SignalReply_Offer{
						Offer: proto.SessionDescriptionFromRTC(*desc),
					},
				})
				if err != nil {
					log.Err(err).Msg("Failed to notify peer about offer.")
					return
				}
			}

			peer.OnIceCandidate = func(init *webrtc.ICECandidateInit, target int) {
				err = stream.Notify(&proto.SignalReply{
					Payload: &proto.SignalReply_Trickle{
						Trickle: &proto.Trickle{
							Target:    int32(target),
							Candidate: proto.CandidateInitFromRTC(*init),
						},
					},
				})
				if err != nil {
					log.Err(err).Msg("Failed to notify peer about trickle.")
					return
				}
			}

			err = peer.Join(payload.Join.SessionId, payload.Join.UserId)
			if err != nil {
				_ = stream.Error(request.Id, err.Error())
				log.Err(err).Msg("Failed to join.")
				return status.Errorf(codes.Internal, fmt.Sprintf("join error: %v", err))
			}

			answer, err := peer.Answer(proto.SessionDescriptionToRTC(payload.Join.Description))
			if err != nil {
				_ = stream.Error(request.Id, err.Error())
				log.Err(err).Msg("Failed to answer.")
				return status.Errorf(codes.Internal, fmt.Sprintf("answer error: %v", err))
			}

			err = stream.Reply(request.Id, &proto.SignalReply{
				Payload: &proto.SignalReply_Join{
					Join: &proto.SessionDescription{
						Type: int32(answer.Type),
						Sdp:  answer.SDP,
					},
				},
			})

			if err != nil {
				log.Err(err).Msg("Failed to reply to join.")
				return status.Errorf(codes.Internal, "join error %s", err)
			}

		case *proto.SignalRequest_Offer:
			answer, err := peer.Answer(proto.SessionDescriptionToRTC(payload.Offer))
			if err != nil {
				_ = stream.Error(request.Id, err.Error())
				log.Err(err).Msg("Failed to answer.")
				return status.Errorf(codes.Internal, fmt.Sprintf("answer error: %v", err))
			}
			err = stream.Reply(request.Id, &proto.SignalReply{
				Payload: &proto.SignalReply_Offer{
					Offer: &proto.SessionDescription{
						Type: int32(answer.Type),
						Sdp:  answer.SDP,
					},
				},
			})
			if err != nil {
				log.Err(err).Msg("Failed to reply to offer.")
				return status.Errorf(codes.Internal, fmt.Sprintf("answer error: %v", err))
			}

		case *proto.SignalRequest_Answer:
			err = peer.SetRemoteDescription(proto.SessionDescriptionToRTC(payload.Answer))
			if err != nil {
				_ = stream.Error(request.Id, err.Error())
				log.Err(err).Msg("Failed to set remote description.")
				return status.Errorf(codes.Internal, fmt.Sprintf("set description error: %v", err))
			}

		case *proto.SignalRequest_Trickle:
			err = peer.Trickle(
				proto.CandidateInitToRTC(payload.Trickle.Candidate),
				int(payload.Trickle.Target),
			)
			if err != nil {
				_ = stream.Error(request.Id, err.Error())
				log.Err(err).Msg("Failed to trickle.")
				return status.Errorf(codes.Internal, fmt.Sprintf("trickle error: %v", err))
			}
		}
	}
}

type stream struct {
	signal proto.SFU_SignalServer
	lock   sync.Mutex
}

func newStream(s proto.SFU_SignalServer) *stream {
	return &stream{signal: s}
}

func (s *stream) Receive() (*proto.SignalRequest, error) {
	return s.signal.Recv()
}

func (s *stream) Notify(rep *proto.SignalReply) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	rep.Id = 0
	return s.signal.Send(rep)
}

func (s *stream) Reply(id uint32, rep *proto.SignalReply) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	rep.Id = id
	return s.signal.Send(rep)
}

func (s *stream) Error(id uint32, msg string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.signal.Send(&proto.SignalReply{
		Id: id,
		Payload: &proto.SignalReply_Error{
			Error: msg,
		},
	})
}

package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	proto "github.com/aromancev/confa/proto/sfu"
)

type SFUServer struct {
	proto.UnimplementedSFUServer
	sync.Mutex
	SFU *sfu.SFU
}

func NewServer(s *sfu.SFU) *SFUServer {
	return &SFUServer{SFU: s}
}

func (s *SFUServer) Signal(stream proto.SFU_SignalServer) error {
	peer := sfu.NewPeer(s.SFU)
	for {
		in, err := stream.Recv()

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

		switch payload := in.Payload.(type) {
		case *proto.SignalRequest_Join:
			var req join
			err := json.Unmarshal(payload.Join, &req)
			if err != nil {
				panic(err)
			}

			peer.OnOffer = func(offer *webrtc.SessionDescription) {
				bytes, err := json.Marshal(offer)
				if err != nil {
					panic(err)
				}
				s.Lock()
				err = stream.Send(&proto.SignalReply{
					Payload: &proto.SignalReply_Description{
						Description: bytes,
					},
				})
				s.Unlock()
				if err != nil {
					panic(err)
				}
			}

			peer.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
				bytes, err := json.Marshal(trickle{
					Candidate: *candidate,
					Target:    target,
				})
				if err != nil {
					panic(err)
				}
				s.Lock()
				err = stream.Send(&proto.SignalReply{
					Payload: &proto.SignalReply_Trickle{
						Trickle: bytes,
					},
				})
				s.Unlock()
				if err != nil {
					panic(err)
				}
			}

			err = peer.Join(req.Sid)
			if err != nil {
				panic(err)
			}

			answer, err := peer.Answer(req.Offer)
			if err != nil {
				return status.Errorf(codes.Internal, fmt.Sprintf("answer error: %v", err))
			}

			bytes, err := json.Marshal(answer)
			if err != nil {
				return status.Errorf(codes.Internal, fmt.Sprintf("sdp marshal error: %v", err))
			}

			s.Lock()
			err = stream.Send(&proto.SignalReply{
				Id: in.Id,
				Payload: &proto.SignalReply_Join{
					Join: bytes,
				},
			})
			s.Unlock()

			if err != nil {
				return status.Errorf(codes.Internal, "join error %s", err)
			}

		case *proto.SignalRequest_Offer:
			var req negotiation
			err := json.Unmarshal(payload.Offer, &req)
			if err != nil {
				panic(err)
			}
			answer, err := peer.Answer(req.Desc)
			if err != nil {
				panic(err)
			}
			bytes, err := json.Marshal(answer)
			if err != nil {
				panic(err)
			}
			s.Lock()
			err = stream.Send(&proto.SignalReply{
				Id: in.Id,
				Payload: &proto.SignalReply_Offer{
					Offer: bytes,
				},
			})
			s.Unlock()
			if err != nil {
				panic(err)
			}

		case *proto.SignalRequest_Answer:
			var req negotiation
			err := json.Unmarshal(payload.Answer, &req)
			if err != nil {
				panic(err)
			}
			err = peer.SetRemoteDescription(req.Desc)
			if err != nil {
				panic(err)
			}

		case *proto.SignalRequest_Trickle:
			var req trickle
			err := json.Unmarshal(payload.Trickle, &req)
			if err != nil {
				panic(err)
			}

			err = peer.Trickle(req.Candidate, req.Target)
			if err != nil {
				panic(err)
			}
		}
	}
}

// join message sent when initializing a peer connection.
type join struct {
	Sid   string                    `json:"sid"`
	Offer webrtc.SessionDescription `json:"offer"`
}

// negotiation message sent when renegotiating the peer connection.
type negotiation struct {
	Desc webrtc.SessionDescription `json:"desc"`
}

// trickle message sent when renegotiating the peer connection.
type trickle struct {
	Target    int                     `json:"target"`
	Candidate webrtc.ICECandidateInit `json:"candidate"`
}

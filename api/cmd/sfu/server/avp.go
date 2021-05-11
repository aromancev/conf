package server

import (
	"context"
	"fmt"
	"time"

	"github.com/pion/ion-avp/pkg/elements"
	"github.com/rs/zerolog/log"

	proto "github.com/aromancev/confa/proto/avp"
	"github.com/aromancev/confa/proto/sfu"

	"github.com/google/uuid"
	avp "github.com/pion/ion-avp/pkg"
	"github.com/pion/webrtc/v3"
	grpcpool "github.com/processout/grpc-go-pool"
)

func init() {
	avp.Init(map[string]avp.ElementFun{
		proto.Process_SAVE.String(): func(sid, pid, tid string, config []byte) avp.Element {
			w := elements.NewFileWriter(
				fmt.Sprintf("/var/lib/media/test.webm", tid),
				4096,
			)
			webm := elements.NewWebmSaver()
			webm.Attach(w)
			return webm
		},
	})
}

type AVP struct {
	proto.UnimplementedAVPServer

	config avp.Config
	pool   *grpcpool.Pool

	peer      *sfu.Peer
	transport *avp.WebRTCTransport
}

func NewAVP(pool *grpcpool.Pool, c avp.Config) *AVP {
	a := &AVP{
		config: c,
		pool:   pool,
	}
	return a
}

func (a *AVP) Signal(ctx context.Context, request *proto.Request) (*proto.Reply, error) {
	pid := uuid.New().String()

	a.transport = avp.NewWebRTCTransport(request.SessionId, a.config)
	offer, err := a.transport.CreateOffer()
	if err != nil {
		panic(err)
	}
	a.peer, err = sfu.NewPeer(context.Background(), a.pool)
	if err != nil {
		panic(err)
	}

	a.peer.OnOffer(func(offer webrtc.SessionDescription) {
		//log.Info().Interface("offer", offer).Msg("Got new offer")
		answer, err := a.transport.Answer(offer)
		if err != nil {
			panic(err)
		}
		err = a.peer.Answer(ctx, answer)
		if err != nil {
			panic(err)
		}
	})
	a.peer.OnTrickle(func(target int, candidate webrtc.ICECandidateInit) {
		err := a.transport.AddICECandidate(candidate, target)
		if err != nil {
			panic(err)
		}
	})

	a.transport.OnICECandidate(func(candidate *webrtc.ICECandidate, target int) {
		if candidate == nil {
			return
		}
		err := a.peer.Trickle(ctx, target, candidate.ToJSON())
		if err != nil {
			panic(err)
		}
	})

	answer, err := a.peer.Join(ctx, request.SessionId, pid, offer)
	if err != nil {
		panic(err)
	}
	err = a.transport.SetRemoteDescription(answer)
	if err != nil {
		panic(err)
	}

	err = a.transport.Process(pid, request.TrackId, request.Process.String(), nil)
	if err != nil {
		panic(err)
	}
	log.Info().Str("track", request.TrackId).Msg("Processed")

	go func() {
		time.Sleep(10 * time.Second)
		_ = a.transport.Close()
		_ = a.peer.Close()
	}()
	return &proto.Reply{}, nil
}

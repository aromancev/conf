package handler

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/pion/ion-avp/pkg/elements"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	proto "github.com/aromancev/confa/proto/media"
	"github.com/aromancev/confa/proto/queue"
	"github.com/aromancev/confa/proto/sfu"

	"github.com/google/uuid"
	avp "github.com/pion/ion-avp/pkg"
	"github.com/pion/webrtc/v3"
	grpcpool "github.com/processout/grpc-go-pool"
)

func InitAVP(mediaDir string, producer Producer) {
	avp.Init(map[string]avp.ElementFun{
		proto.Process_SAVE.String(): func(sid, pid, tid string, config []byte) avp.Element {
			log.Info().Msg("Save process started.")
			err := os.MkdirAll(path.Join(mediaDir, tid), 0777)
			if err != nil {
				log.Err(err).Msg("Failed to create directory.")
				return nil
			}
			webm := elements.NewWebmSaver()
			webm.Attach(elements.NewFileWriter(path.Join(mediaDir, tid, "raw.webm"), 4096))
			webm.Attach(newProgressElement(tid, producer))
			return webm
		},
	})
}

type AVP struct {
	proto.UnimplementedAVPServer

	config avp.Config
	pool   *grpcpool.Pool

	m     sync.Mutex
	procs map[string]*process
}

func NewAVP(c avp.Config, pool *grpcpool.Pool) *AVP {
	a := &AVP{
		config: c,
		pool:   pool,
		procs:  make(map[string]*process),
	}
	return a
}

func (a *AVP) Signal(_ context.Context, request *proto.Request) (*proto.Reply, error) {
	proc := a.proc(request.SessionId)
	err := proc.Start(context.Background(), a.pool, request.TrackId, request.Process, nil)
	if err != nil {
		log.Err(err).Msg("Failed to start processing.")
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	log.Info().Str("session", request.SessionId).Str("track", request.TrackId).Msg("Processing started.")

	uid, _ := proc.id.MarshalBinary()
	return &proto.Reply{Uid: uid}, nil
}

func (a *AVP) proc(sessionID string) *process {
	a.m.Lock()
	defer a.m.Unlock()

	proc, ok := a.procs[sessionID]
	if ok {
		return proc
	}
	proc = newProcess(sessionID, a.config, func() {
		a.m.Lock()
		defer a.m.Unlock()

		delete(a.procs, sessionID)
		log.Info().Str("session", sessionID).Msg("Closed process.")
	})
	a.procs[sessionID] = proc
	log.Info().Str("session", sessionID).Msg("Created new process.")
	return proc
}

type process struct {
	id        uuid.UUID
	sessionID string
	onClose   func()

	m         sync.Mutex
	peer      *sfu.Peer
	transport *avp.WebRTCTransport
}

func newProcess(sessionID string, config avp.Config, onClose func()) *process {
	return &process{
		id:        uuid.New(),
		sessionID: sessionID,
		onClose:   onClose,
		transport: avp.NewWebRTCTransport(sessionID, config),
	}
}

func (p *process) Start(ctx context.Context, pool *grpcpool.Pool, trackID string, process proto.Process, config []byte) error {
	if err := p.connect(ctx, pool); err != nil {
		return err
	}

	err := p.transport.Process(p.id.String(), trackID, process.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}
	return nil
}

func (p *process) connect(ctx context.Context, pool *grpcpool.Pool) error {
	p.m.Lock()
	defer p.m.Unlock()

	if p.peer != nil {
		return nil
	}

	var err error
	p.peer, err = sfu.NewPeer(ctx, pool)
	if err != nil {
		return fmt.Errorf("failed to create peer: %w", err)
	}
	p.peer.OnOffer(func(offer webrtc.SessionDescription) {
		answer, err := p.transport.Answer(offer)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to create answer.")
			return
		}
		err = p.peer.Answer(ctx, answer)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to answer peer.")
			return
		}
	})
	p.peer.OnTrickle(func(target int, candidate webrtc.ICECandidateInit) {
		err := p.transport.AddICECandidate(candidate, target)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to add ICE Candidate.")
			return
		}
	})
	p.transport.OnICECandidate(func(candidate *webrtc.ICECandidate, target int) {
		if candidate == nil {
			return
		}
		err := p.peer.Trickle(ctx, target, candidate.ToJSON())
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to send trickle to peer.")
			return
		}
	})
	p.transport.OnClose(func() {
		p.m.Lock()
		defer p.m.Unlock()

		_ = p.peer.Close()
		p.onClose()
		log.Info().Msg("Transport closed.")
	})

	offer, err := p.transport.CreateOffer()
	if err != nil {
		return fmt.Errorf("failed to create offer: %w", err)
	}
	answer, err := p.peer.Join(ctx, p.sessionID, p.id.String(), offer)
	if err != nil {
		return fmt.Errorf("failed to join: %w", err)
	}
	err = p.transport.SetRemoteDescription(answer)
	if err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}
	return nil
}

type progressElement struct {
	mediaID  string
	producer Producer
}

func newProgressElement(mediaID string, producer Producer) *progressElement {
	return &progressElement{mediaID: mediaID, producer: producer}
}

func (e *progressElement) Write(_ *avp.Sample) error {
	return nil
}

func (e *progressElement) Attach(_ avp.Element) {
}

func (e *progressElement) Close() {
	body, err := queue.Marshal(&queue.VideoJob{MediaId: e.mediaID}, "TODO")
	if err != nil {
		log.Err(err).Msg("Failed to marshal video processing job.")
		return
	}
	_, err = e.producer.Put(context.Background(), queue.TubeVideo, body, beanstalk.PutParams{})
	if err != nil {
		log.Err(err).Msg("Failed to put video job.")
		return
	}
}

package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/event/proxy"
	"github.com/aromancev/confa/internal/platform/signal"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var (
	ErrClosed = errors.New("connection closed")
)

type PeerProxy struct {
	proxy   *proxy.Proxy
	conn    *websocket.Conn
	sfuConn *grpc.ClientConn
	workers *workers
	signal  *signal.GRPCSignal
}

func NewPeerProxy(ctx context.Context, w http.ResponseWriter, r *http.Request, userID, roomID uuid.UUID, watcher event.Watcher, emitter proxy.EventEmitter, sfuConn *grpc.ClientConn) (*PeerProxy, error) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to accept websocket connection: %w", err)
	}

	cursor, err := watcher.Watch(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return &PeerProxy{
		conn:    conn,
		sfuConn: sfuConn,
		proxy:   proxy.NewProxy(ctx, userID, roomID, cursor, emitter),
		workers: newWorkers(ctx),
	}, nil
}

func (p *PeerProxy) Serve() {
	p.workers.Serve(func(ctx context.Context) {
		for {
			err := p.receiveWebsocket(ctx)
			switch {
			case errors.Is(err, proxy.ErrValidation):
				log.Ctx(ctx).Warn().Err(err).Msg("Message from websocket rejected.")
				continue
			case errors.Is(err, ErrClosed), errors.Is(err, io.EOF):
				return
			case err != nil:
				log.Ctx(ctx).Err(err).Msg("Failed to process websocket message.")
				return
			}
		}
	})
	p.workers.Serve(func(ctx context.Context) {
		for {
			err := p.ping(ctx)
			switch {
			case errors.Is(err, io.EOF), errors.Is(err, ErrClosed), errors.Is(err, context.DeadlineExceeded):
				return
			case err != nil:
				log.Ctx(ctx).Err(err).Msg("Websocket ping failed.")
				return
			}
		}
	})
	p.workers.Serve(func(ctx context.Context) {
		for {
			ev, err := p.proxy.RecieveEvent(ctx)
			switch {
			case errors.Is(err, proxy.ErrUnknownMessage):
				log.Ctx(ctx).Debug().Msg("Skipping unknown event.")
				continue
			case errors.Is(err, context.Canceled):
				return
			case err != nil:
				log.Ctx(ctx).Err(err).Msg("Failed to receive event.")
				return
			}
			err = wsjson.Write(ctx, p.conn, Message{
				Payload: MessagePayload{
					Event: NewRoomEvent(ev),
				},
			})
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to reply to websocket.")
				return
			}
		}
	})
	p.workers.Wait()
}

func (p *PeerProxy) Close(ctx context.Context) {
	p.workers.Close()
	p.proxy.Close(ctx)
	_ = p.conn.Close(websocket.StatusNormalClosure, "Peer closed.")
	if p.signal != nil {
		_ = p.signal.Close(ctx)
	}
}

func (p *PeerProxy) serveSignal(ctx context.Context) {
	for {
		msg, err := p.proxy.ReceiveSignal(ctx, p.signal)
		switch {
		case errors.Is(err, proxy.ErrUnknownMessage):
			log.Ctx(ctx).Debug().Msg("Skipping unknown signal.")
			continue
		case errors.Is(err, proxy.ErrClosed), errors.Is(err, context.Canceled):
			log.Ctx(ctx).Debug().Msg("Serving signal cancelled.")
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to receive signal.")
			return
		}
		err = wsjson.Write(ctx, p.conn, Message{
			Payload: MessagePayload{
				Signal: newSignal(msg),
			},
		})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to reply to websocket.")
			return
		}
	}
}

func (p *PeerProxy) receiveWebsocket(ctx context.Context) error {
	var msg Message
	err := wsjson.Read(ctx, p.conn, &msg)
	switch {
	case errors.Is(err, context.Canceled), errors.As(err, &websocket.CloseError{}):
		return ErrClosed
	case err != nil:
		return err
	}

	switch {
	case msg.Payload.Signal != nil:
		pl := *msg.Payload.Signal
		if p.signal == nil {
			p.signal, err = signal.NewGRPCSignal(ctx, p.sfuConn)
			if err != nil {
				return fmt.Errorf("failed to connect to signal: %w", err)
			}
			p.workers.Serve(p.serveSignal)
		}
		return p.proxy.SendSignal(ctx, p.signal, signalMessage(pl))
	case msg.Payload.State != nil:
		pl := *msg.Payload.State
		state, err := p.proxy.SendState(ctx, peerState(pl))
		if err != nil {
			return err
		}
		schemaState := PeerState{
			Tracks: make([]Track, len(state.Tracks)),
		}
		for i, t := range state.Tracks {
			schemaState.Tracks[i] = Track{
				ID:   t.ID,
				Hint: Hint(t.Hint),
			}
		}
		return wsjson.Write(ctx, p.conn, Message{
			ResponseID: msg.RequestID,
			Payload: MessagePayload{
				State: &schemaState,
			},
		})
	case msg.Payload.PeerMessage != nil:
		pl := *msg.Payload.PeerMessage
		ev, err := p.proxy.SendMessage(ctx, pl.Text)
		if err != nil {
			return err
		}
		return wsjson.Write(ctx, p.conn, Message{
			ResponseID: msg.RequestID,
			Payload: MessagePayload{
				Event: NewRoomEvent(ev),
			},
		})
	}
	log.Ctx(ctx).Debug().Msg("Skipping unknown message.")
	return nil
}

func (p *PeerProxy) ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	err := p.conn.Ping(pingCtx)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ErrClosed
	case <-time.After(20 * time.Second):
		return nil
	}
}

func newSignal(msg signal.Message) *Signal {
	var s Signal
	switch {
	case msg.Join != nil:
		pl := msg.Join
		s.Join = &SignalJoin{
			UserID:      pl.UserID,
			SessionID:   pl.SessionID,
			Description: newSessionDescription(pl.Description),
		}
	case msg.Offer != nil:
		pl := msg.Offer
		s.Offer = &SignalOffer{
			Description: newSessionDescription(pl.Description),
		}
	case msg.Answer != nil:
		pl := msg.Answer
		s.Answer = &SignalAnswer{
			Description: newSessionDescription(pl.Description),
		}
	case msg.Trickle != nil:
		pl := msg.Trickle
		s.Trickle = &SignalTrickle{
			Candidate: newICECandidateInit(pl.Candidate),
			Target:    int64(pl.Target),
		}
	}
	return &s
}

func signalMessage(s Signal) signal.Message {
	var msg signal.Message
	switch {
	case s.Join != nil:
		pl := s.Join
		msg.Join = &signal.Join{
			UserID:      pl.UserID,
			SessionID:   pl.SessionID,
			Description: webrtcSessionDescription(pl.Description),
		}
	case s.Offer != nil:
		pl := s.Offer
		msg.Offer = &signal.Offer{
			Description: webrtcSessionDescription(pl.Description),
		}
	case s.Answer != nil:
		pl := s.Answer
		msg.Answer = &signal.Answer{
			Description: webrtcSessionDescription(pl.Description),
		}
	case s.Trickle != nil:
		pl := s.Trickle
		msg.Trickle = &signal.Trickle{
			Candidate: webrtcICECandidateInit(pl.Candidate),
			Target:    int(pl.Target),
		}
	}
	return msg
}

func newSessionDescription(d webrtc.SessionDescription) SessionDescription {
	return SessionDescription{
		Type: SDPType(d.Type.String()),
		SDP:  d.SDP,
	}
}

func webrtcSessionDescription(d SessionDescription) webrtc.SessionDescription {
	return webrtc.SessionDescription{
		Type: webrtc.NewSDPType(string(d.Type)),
		SDP:  d.SDP,
	}
}

func newICECandidateInit(webrtcInit webrtc.ICECandidateInit) ICECandidateInit {
	init := ICECandidateInit{
		Candidate:        webrtcInit.Candidate,
		SDPMid:           webrtcInit.SDPMid,
		UsernameFragment: webrtcInit.UsernameFragment,
	}
	if webrtcInit.SDPMLineIndex != nil {
		index := int64(*webrtcInit.SDPMLineIndex)
		init.SDPMLineIndex = &index
	}
	return init
}

func webrtcICECandidateInit(init ICECandidateInit) webrtc.ICECandidateInit {
	webrtcInit := webrtc.ICECandidateInit{
		Candidate:        init.Candidate,
		SDPMid:           init.SDPMid,
		UsernameFragment: init.UsernameFragment,
	}
	if init.SDPMLineIndex != nil {
		index := uint16(*init.SDPMLineIndex)
		webrtcInit.SDPMLineIndex = &index
	}
	return webrtcInit
}

func peerState(state PeerState) proxy.State {
	tracks := make([]event.Track, len(state.Tracks))
	for i, t := range state.Tracks {
		tracks[i] = event.Track{
			ID:   t.ID,
			Hint: event.TrackHint(t.Hint),
		}
	}
	return proxy.State{
		Tracks: tracks,
	}
}

type workers struct {
	done   chan struct{}
	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup
}

func newWorkers(ctx context.Context) *workers {
	ctx, cancel := context.WithCancel(ctx)
	return &workers{
		done:   make(chan struct{}),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (w *workers) Serve(f func(context.Context)) {
	defer func() {
		if err := recover(); err != nil {
			log.Ctx(w.ctx).Error().Interface("err", err).Msg("Failed to join worker pool.")
		}
	}()

	w.wg.Add(1)
	go func() {
		f(w.ctx)
		w.cancel()
		w.wg.Done()
	}()
}

func (w *workers) Wait() {
	w.wg.Wait()
}

func (w *workers) Close() {
	w.cancel()
}

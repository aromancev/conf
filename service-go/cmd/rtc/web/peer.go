package web

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/event/peer"
	"github.com/aromancev/confa/internal/platform/signal"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var (
	ErrClosed = errors.New("connection closed")
)

type Peer struct {
	peer    *peer.Peer
	conn    *safeConn
	sfuConn *grpc.ClientConn
}

func NewPeer(ctx context.Context, w http.ResponseWriter, r *http.Request, userID, roomID uuid.UUID, watcher event.Watcher, emitter peer.EventEmitter, sfuConn *grpc.ClientConn) (*Peer, error) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	cursor, err := watcher.Watch(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return &Peer{
		conn:    &safeConn{conn: conn},
		sfuConn: sfuConn,
		peer:    peer.NewPeer(ctx, userID, roomID, cursor, emitter),
	}, nil
}

func (p *Peer) Serve(ctx context.Context, connectMedia bool) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var signalClient *signal.GRPCSignal
	var wg sync.WaitGroup
	wg.Add(2)
	if connectMedia {
		var err error
		signalClient, err = signal.NewGRPCSignal(ctx, p.sfuConn)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to connect to signal.")
			return
		}
		defer signalClient.Close(ctx)

		wg.Add(1)
		go func() {
			p.serveSignal(ctx, signalClient)
			cancel()
			wg.Done()
		}()
	}
	go func() {
		p.serveWebsocket(ctx, signalClient)
		cancel()
		wg.Done()
	}()
	go func() {
		p.serveEvents(ctx)
		cancel()
		wg.Done()
	}()
	wg.Wait()
}

func (p *Peer) Close(ctx context.Context) {
	p.peer.Close(ctx)
	p.conn.Close()
}

func (p *Peer) serveWebsocket(ctx context.Context, sig peer.Signal) {
	for {
		err := p.receiveWebsocket(ctx, sig)
		switch {
		case errors.Is(err, peer.ErrValidation):
			log.Ctx(ctx).Warn().Err(err).Msg("Message from websocket rejected.")
			continue
		case errors.Is(err, ErrClosed):
			log.Ctx(ctx).Info().Msg("Websocket disconnected.")
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to process message.")
			return
		}
	}
}

func (p *Peer) serveSignal(ctx context.Context, sig peer.Signal) {
	if sig == nil {
		panic("signal client not provided")
	}
	for {
		msg, err := p.peer.ReceiveSignal(ctx, sig)
		switch {
		case errors.Is(err, peer.ErrUnknownMessage):
			log.Ctx(ctx).Debug().Msg("Skipping unknown signal.")
			continue
		case errors.Is(err, peer.ErrClosed), errors.Is(err, context.Canceled):
			log.Ctx(ctx).Debug().Msg("Serving signal cancelled.")
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to receive signal.")
			return
		}
		err = p.conn.WriteJSON(Message{
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

func (p *Peer) serveEvents(ctx context.Context) {
	for {
		ev, err := p.peer.RecieveEvent(ctx)
		switch {
		case errors.Is(err, peer.ErrUnknownMessage):
			log.Ctx(ctx).Debug().Msg("Skipping unknown event.")
			continue
		case errors.Is(err, context.Canceled):
			log.Ctx(ctx).Debug().Msg("Serving events cancelled.")
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to receive event.")
			return
		}
		err = p.conn.WriteJSON(Message{
			Payload: MessagePayload{
				Event: NewRoomEvent(ev),
			},
		})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to reply to websocket.")
			return
		}
	}
}

func (p *Peer) receiveWebsocket(ctx context.Context, sig peer.Signal) error {
	var msg Message
	err := p.conn.ReadJSON(&msg)
	switch {
	case websocket.IsCloseError(err, websocket.CloseGoingAway), errors.Is(err, context.Canceled):
		return ErrClosed
	case err != nil:
		return err
	}

	switch {
	case msg.Payload.Signal != nil && sig != nil:
		pl := *msg.Payload.Signal
		return p.peer.SendSignal(ctx, sig, signalMessage(pl))
	case msg.Payload.State != nil:
		pl := *msg.Payload.State
		state, err := p.peer.SendState(ctx, peerState(pl))
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
		return p.conn.WriteJSON(Message{
			ResponseID: msg.RequestID,
			Payload: MessagePayload{
				State: &schemaState,
			},
		})
	case msg.Payload.PeerMessage != nil:
		pl := *msg.Payload.PeerMessage
		ev, err := p.peer.SendMessage(ctx, pl.Text)
		if err != nil {
			return err
		}
		return p.conn.WriteJSON(Message{
			ResponseID: msg.RequestID,
			Payload: MessagePayload{
				Event: NewRoomEvent(ev),
			},
		})
	}
	log.Ctx(ctx).Debug().Msg("Skipping unknown message.")
	return nil
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

func peerState(state PeerState) peer.State {
	tracks := make([]event.Track, len(state.Tracks))
	for i, t := range state.Tracks {
		tracks[i] = event.Track{
			ID:   t.ID,
			Hint: event.TrackHint(t.Hint),
		}
	}
	return peer.State{
		Tracks: tracks,
	}
}

// TODO: Use https://github.com/nhooyr/websocket
type safeConn struct {
	l    sync.Mutex
	conn *websocket.Conn
}

func (c *safeConn) ReadJSON(v interface{}) error {
	return c.conn.ReadJSON(v)
}

func (c *safeConn) WriteJSON(v interface{}) error {
	c.l.Lock()
	defer c.l.Unlock()
	return c.conn.WriteJSON(v)
}

func (c *safeConn) Close() {
	c.conn.Close()
}

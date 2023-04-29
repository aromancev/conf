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
	"github.com/aromancev/confa/event/peer"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var (
	ErrClosed = errors.New("connection closed")
)

type SocketPeer struct {
	peer    *peer.Peer
	conn    *websocket.Conn
	workers *workers
}

func NewSocketPeer(ctx context.Context, w http.ResponseWriter, r *http.Request, userID, roomID uuid.UUID, watcher event.Watcher, emitter peer.EventEmitter) (*SocketPeer, error) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to accept websocket connection: %w", err)
	}

	cursor, err := watcher.Watch(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return &SocketPeer{
		conn:    conn,
		peer:    peer.NewPeer(ctx, userID, roomID, cursor, emitter),
		workers: newWorkers(ctx),
	}, nil
}

func (p *SocketPeer) SessionID() uuid.UUID {
	return p.peer.SessionID()
}

func (p *SocketPeer) Serve() {
	p.workers.Serve(func(ctx context.Context) {
		for {
			err := p.receiveWebsocket(ctx)
			switch {
			case errors.Is(err, peer.ErrValidation):
				log.Ctx(ctx).Warn().Err(err).Msg("Message from websocket rejected.")
				continue
			case errors.Is(err, ErrClosed), errors.Is(err, io.EOF), errors.As(err, &websocket.CloseError{}):
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
			case errors.Is(err, io.EOF), errors.Is(err, ErrClosed), errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled), errors.As(err, &websocket.CloseError{}):
				return
			case err != nil:
				log.Ctx(ctx).Err(err).Msg("Websocket ping failed.")
				return
			}
		}
	})
	p.workers.Serve(func(ctx context.Context) {
		for {
			ev, err := p.peer.RecieveEvent(ctx)
			switch {
			case errors.Is(err, peer.ErrUnknownMessage):
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
			switch {
			case errors.As(err, &websocket.CloseError{}):
				return
			case err != nil:
				log.Ctx(ctx).Err(err).Msg("Failed to reply to websocket.")
				return
			}
		}
	})
	p.workers.Wait()
}

func (p *SocketPeer) Close(ctx context.Context) {
	p.workers.Close()
	p.peer.Close(ctx)
	_ = p.conn.Close(websocket.StatusNormalClosure, "Peer closed.")
}

func (p *SocketPeer) receiveWebsocket(ctx context.Context) error {
	var msg Message
	err := wsjson.Read(ctx, p.conn, &msg)
	switch {
	case errors.Is(err, context.Canceled), errors.As(err, &websocket.CloseError{}):
		return ErrClosed
	case err != nil:
		return err
	}

	switch {
	case msg.Payload.PeerMessage != nil:
		pl := *msg.Payload.PeerMessage
		ev, err := p.peer.SendMessage(ctx, pl.Text)
		if err != nil {
			return err
		}
		return wsjson.Write(ctx, p.conn, Message{
			ResponseID: msg.RequestID,
			Payload: MessagePayload{
				Event: NewRoomEvent(ev),
			},
		})
	case msg.Payload.Reaction != nil:
		pl := *msg.Payload.Reaction
		var react event.Reaction
		if pl.Clap != nil {
			react.Clap = &event.ReactionClap{
				IsStarting: pl.Clap.IsStarting,
			}
		}
		ev, err := p.peer.SendReaction(ctx, react)
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

func (p *SocketPeer) ping(ctx context.Context) error {
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

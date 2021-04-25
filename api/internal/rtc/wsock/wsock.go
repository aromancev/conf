package wsock

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var (
	ErrClosed = errors.New("websocket closed")
)

type Upgrader struct {
	upgrader websocket.Upgrader
}

func NewUpgrader(readBuf, writeBuf int) *Upgrader {
	return &Upgrader{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  readBuf,
			WriteBufferSize: writeBuf,
		},
	}
}

func (u Upgrader) NewConn(w http.ResponseWriter, r *http.Request, h http.Header) (*Conn, error) {
	conn, err := u.upgrader.Upgrade(w, r, h)
	if err != nil {
		return nil, err
	}
	return &Conn{conn: conn}, nil
}

type Conn struct {
	lock sync.Mutex
	conn *websocket.Conn
}

func (c *Conn) Receive() (interface{}, error) {
	var req request
	err := c.conn.ReadJSON(&req)
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			return nil, ErrClosed
		}
		return nil, err
	}
	switch req.Type {
	case typeJoin:
		join := Join{
			socketRequest: newSocketRequest(c, req.ID),
		}
		err := json.Unmarshal(req.Payload, &join)
		if err != nil {
			return nil, err
		}
		return join, nil

	case typeOffer:
		offer := Offer{
			socketRequest: newSocketRequest(c, req.ID),
		}
		err := json.Unmarshal(req.Payload, &offer.Offer)
		if err != nil {
			return nil, err
		}
		return offer, nil

	case typeAnswer:
		answer := Answer{
			socketRequest: newSocketRequest(c, req.ID),
		}
		err := json.Unmarshal(req.Payload, &answer.Answer)
		if err != nil {
			return nil, err
		}
		return answer, nil

	case typeTrickle:
		trickle := Trickle{
			socketRequest: newSocketRequest(c, req.ID),
		}
		err := json.Unmarshal(req.Payload, &trickle)
		if err != nil {
			return nil, err
		}
		return trickle, nil

	default:
		return nil, errors.New("unknown request type")
	}
}

func (c *Conn) NotifyOffer(offer webrtc.SessionDescription) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.conn.WriteJSON(response{
		Type:    typeOffer,
		Payload: offer,
	})
}

func (c *Conn) NotifyTrickle(target int32, candidate webrtc.ICECandidateInit) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.conn.WriteJSON(response{
		Type: typeTrickle,
		Payload: Trickle{
			Target:    target,
			Candidate: candidate,
		},
	})
}

func (c *Conn) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.conn.Close()
}

type Join struct {
	socketRequest

	Sid   string                    `json:"sid"`
	Offer webrtc.SessionDescription `json:"offer"`
}

func (r Join) Reply(answer webrtc.SessionDescription) error {
	return r.socket.response(response{
		ID:      r.id,
		Type:    typeOffer,
		Payload: answer,
	})
}

type Offer struct {
	socketRequest

	Offer webrtc.SessionDescription
}

func (r Offer) Reply(answer webrtc.SessionDescription) error {
	return r.socket.response(response{
		ID:      r.id,
		Type:    typeOffer,
		Payload: answer,
	})
}

type Trickle struct {
	socketRequest

	Target    int32                   `json:"target"`
	Candidate webrtc.ICECandidateInit `json:"candidate"`
}

type Answer struct {
	socketRequest

	Answer webrtc.SessionDescription
}

func (c *Conn) response(resp response) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.conn.WriteJSON(resp)
}

type socketRequest struct {
	id     uint32
	socket *Conn
}

func newSocketRequest(s *Conn, id uint32) socketRequest {
	return socketRequest{id: id, socket: s}
}

func (r socketRequest) Error(err string) error {
	return r.socket.response(response{
		ID:      r.id,
		Type:    typeError,
		Payload: err,
	})
}

const (
	typeJoin    = "join"
	typeOffer   = "offer"
	typeAnswer  = "answer"
	typeTrickle = "trickle"
	typeError   = "error"
)

type request struct {
	ID      uint32          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type response struct {
	ID      uint32      `json:"id,omitempty"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

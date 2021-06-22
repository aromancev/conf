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
		var join Join
		err := json.Unmarshal(req.Payload, &join)
		if err != nil {
			return nil, err
		}
		return join, nil

	case typeOffer:
		var offer Offer
		err := json.Unmarshal(req.Payload, &offer.Description)
		if err != nil {
			return nil, err
		}
		return offer, nil

	case typeAnswer:
		var answer Answer
		err := json.Unmarshal(req.Payload, &answer.Description)
		if err != nil {
			return nil, err
		}
		return answer, nil

	case typeTrickle:
		var trickle Trickle
		err := json.Unmarshal(req.Payload, &trickle)
		if err != nil {
			return nil, err
		}
		return trickle, nil

	default:
		return nil, errors.New("unknown request type")
	}
}

func (c *Conn) Answer(answer webrtc.SessionDescription) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.conn.WriteJSON(response{
		Type:    typeAnswer,
		Payload: answer,
	})
}

func (c *Conn) Offer(offer webrtc.SessionDescription) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.conn.WriteJSON(response{
		Type:    typeOffer,
		Payload: offer,
	})
}

func (c *Conn) Trickle(candidate webrtc.ICECandidateInit, target int) error {
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

func (c *Conn) Error(err string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.conn.WriteJSON(response{
		Type:    typeError,
		Payload: err,
	})
}

func (c *Conn) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.conn.Close()
}

type Join struct {
	SID   string                    `json:"sid"`
	UID   string                    `json:"uid"`
	Offer webrtc.SessionDescription `json:"offer"`
}

type Offer struct {
	Description webrtc.SessionDescription
}

type Trickle struct {
	Target    int                     `json:"target"`
	Candidate webrtc.ICECandidateInit `json:"candidate"`
}

type Answer struct {
	Description webrtc.SessionDescription
}

const (
	typeJoin    = "join"
	typeOffer   = "offer"
	typeAnswer  = "answer"
	typeTrickle = "trickle"
	typeError   = "error"
)

type request struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type response struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

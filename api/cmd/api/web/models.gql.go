// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package web

import (
	"fmt"
	"io"
	"strconv"
)

type ClapInput struct {
	SpeakerID *string `json:"speakerId"`
	ConfaID   *string `json:"confaId"`
	TalkID    *string `json:"talkId"`
}

type Claps struct {
	Value     int `json:"value"`
	UserValue int `json:"userValue"`
}

type Confa struct {
	ID          string `json:"id"`
	OwnerID     string `json:"ownerId"`
	Handle      string `json:"handle"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ConfaInput struct {
	ID          *string `json:"id"`
	OwnerID     *string `json:"ownerId"`
	Handle      *string `json:"handle"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type Confas struct {
	Items    []*Confa `json:"items"`
	Limit    int      `json:"limit"`
	NextFrom string   `json:"nextFrom"`
}

type Event struct {
	ID        string        `json:"id"`
	OwnerID   string        `json:"ownerId"`
	RoomID    string        `json:"roomId"`
	CreatedAt string        `json:"createdAt"`
	Payload   *EventPayload `json:"payload"`
}

type EventFrom struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
}

type EventFromInput struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
}

type EventInput struct {
	RoomID string `json:"roomId"`
}

type EventLimit struct {
	Count   int `json:"count"`
	Seconds int `json:"seconds"`
}

type EventPayload struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type Events struct {
	Items    []*Event   `json:"items"`
	Limit    int        `json:"limit"`
	NextFrom *EventFrom `json:"nextFrom"`
}

type Talk struct {
	ID        string `json:"id"`
	OwnerID   string `json:"ownerId"`
	SpeakerID string `json:"speakerId"`
	ConfaID   string `json:"confaId"`
	RoomID    string `json:"roomId"`
	Handle    string `json:"handle"`
}

type TalkInput struct {
	ID        *string `json:"id"`
	OwnerID   *string `json:"ownerId"`
	SpeakerID *string `json:"speakerId"`
	ConfaID   *string `json:"confaId"`
	Handle    *string `json:"handle"`
}

type Talks struct {
	Items    []*Talk `json:"items"`
	Limit    int     `json:"limit"`
	NextFrom string  `json:"nextFrom"`
}

type Token struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn"`
}

type EventOrder string

const (
	EventOrderAsc  EventOrder = "ASC"
	EventOrderDesc EventOrder = "DESC"
)

var AllEventOrder = []EventOrder{
	EventOrderAsc,
	EventOrderDesc,
}

func (e EventOrder) IsValid() bool {
	switch e {
	case EventOrderAsc, EventOrderDesc:
		return true
	}
	return false
}

func (e EventOrder) String() string {
	return string(e)
}

func (e *EventOrder) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = EventOrder(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid EventOrder", str)
	}
	return nil
}

func (e EventOrder) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

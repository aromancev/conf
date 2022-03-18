// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    message, err := UnmarshalMessage(bytes)
//    bytes, err = message.Marshal()

package web

import "encoding/json"

func UnmarshalMessage(data []byte) (Message, error) {
	var r Message
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Message) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Message struct {
	Payload    MessagePayload `json:"payload"`             
	RequestID  *string        `json:"requestId,omitempty"` 
	ResponseID *string        `json:"responseId,omitempty"`
}

type MessagePayload struct {
	Event       *RoomEvent   `json:"event,omitempty"`      
	PeerMessage *PeerMessage `json:"peerMessage,omitempty"`
	Signal      *Signal      `json:"signal,omitempty"`     
	State       *PeerState   `json:"state,omitempty"`      
}

type RoomEvent struct {
	CreatedAt float64      `json:"createdAt"`
	ID        string       `json:"id"`       
	OwnerID   string       `json:"ownerId"`  
	Payload   EventPayload `json:"payload"`  
	RoomID    string       `json:"roomId"`   
}

type EventPayload struct {
	Message   *EventMessage   `json:"message,omitempty"`  
	PeerState *EventPeerState `json:"peerState,omitempty"`
}

type EventMessage struct {
	Text string `json:"text"`
}

type EventPeerState struct {
	Status *Status `json:"status,omitempty"`
	Tracks []Track `json:"tracks,omitempty"`
}

type Track struct {
	Hint Hint   `json:"hint"`
	ID   string `json:"id"`  
}

type PeerMessage struct {
	Text string `json:"text"`
}

type Signal struct {
	Answer  *SignalAnswer  `json:"answer,omitempty"` 
	Join    *SignalJoin    `json:"join,omitempty"`   
	Offer   *SignalOffer   `json:"offer,omitempty"`  
	Trickle *SignalTrickle `json:"trickle,omitempty"`
}

type SignalAnswer struct {
	Description SessionDescription `json:"description"`
}

type SessionDescription struct {
	SDP  string  `json:"sdp"` 
	Type SDPType `json:"type"`
}

type SignalJoin struct {
	Description SessionDescription `json:"description"`
	SessionID   string             `json:"sessionId"`  
	UserID      string             `json:"userId"`     
}

type SignalOffer struct {
	Description SessionDescription `json:"description"`
}

type SignalTrickle struct {
	Candidate ICECandidateInit `json:"candidate"`
	Target    int64            `json:"target"`   
}

type ICECandidateInit struct {
	Candidate        string  `json:"candidate"`                 
	SDPMid           *string `json:"sdpMid,omitempty"`          
	SDPMLineIndex    *int64  `json:"sdpMLineIndex,omitempty"`   
	UsernameFragment *string `json:"usernameFragment,omitempty"`
}

type PeerState struct {
	Tracks []Track `json:"tracks,omitempty"`
}

type Status string
const (
	Joined Status = "joined"
	Left Status = "left"
)

type Hint string
const (
	Camera Hint = "camera"
	DeviceAudio Hint = "device_audio"
	Screen Hint = "screen"
	UserAudio Hint = "user_audio"
)

type SDPType string
const (
	Answer SDPType = "answer"
	Offer SDPType = "offer"
	Pranswer SDPType = "pranswer"
	Rollback SDPType = "rollback"
)

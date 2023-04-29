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
	Reaction    *Reaction    `json:"reaction,omitempty"`   
}

type RoomEvent struct {
	CreatedAt float64      `json:"createdAt"`
	ID        string       `json:"id"`       
	Payload   EventPayload `json:"payload"`  
	RoomID    string       `json:"roomId"`   
}

type EventPayload struct {
	Message     *EventMessage     `json:"message,omitempty"`    
	PeerState   *EventPeerState   `json:"peerState,omitempty"`  
	Reaction    *EventReaction    `json:"reaction,omitempty"`   
	Recording   *EventRecording   `json:"recording,omitempty"`  
	TrackRecord *EventTrackRecord `json:"trackRecord,omitempty"`
}

type EventMessage struct {
	FromID string `json:"fromId"`
	Text   string `json:"text"`  
}

type EventPeerState struct {
	PeerID    string      `json:"peerId"`          
	SessionID string      `json:"sessionId"`       
	Status    *PeerStatus `json:"status,omitempty"`
}

type EventReaction struct {
	FromID   string   `json:"fromId"`  
	Reaction Reaction `json:"reaction"`
}

type Reaction struct {
	Clap *ReactionClap `json:"clap,omitempty"`
}

type ReactionClap struct {
	IsStarting bool `json:"isStarting"`
}

type EventRecording struct {
	Status RecordingEventStatus `json:"status"`
}

type EventTrackRecord struct {
	Kind     TrackKind   `json:"kind"`    
	RecordID string      `json:"recordId"`
	Source   TrackSource `json:"source"`  
}

type PeerMessage struct {
	Text string `json:"text"`
}

type PeerStatus string
const (
	Joined PeerStatus = "JOINED"
	Left PeerStatus = "LEFT"
)

type RecordingEventStatus string
const (
	Started RecordingEventStatus = "STARTED"
	Stopped RecordingEventStatus = "STOPPED"
)

type TrackKind string
const (
	Audio TrackKind = "AUDIO"
	Video TrackKind = "VIDEO"
)

type TrackSource string
const (
	Camera TrackSource = "CAMERA"
	Microphone TrackSource = "MICROPHONE"
	Screen TrackSource = "SCREEN"
	ScreenAudio TrackSource = "SCREEN_AUDIO"
	Unknown TrackSource = "UNKNOWN"
)

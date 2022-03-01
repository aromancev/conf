package event

import (
	"time"

	"github.com/aromancev/confa/internal/proto/rtc"
	"github.com/google/uuid"
)

func ToProto(event Event) *rtc.Event {
	id, _ := event.ID.MarshalBinary()
	owner, _ := event.Owner.MarshalBinary()
	room, _ := event.Room.MarshalBinary()
	var payload rtc.Event_Payload
	if event.Payload.Message != nil {
		pl := *event.Payload.Message
		payload.Message = &rtc.Event_Payload_PayloadMessage{
			Text: pl.Text,
		}
	}
	if event.Payload.PeerState != nil {
		pl := *event.Payload.PeerState
		payload.PeerState = &rtc.Event_Payload_PayloadPeerState{}
		if pl.Status != "" {
			payload.PeerState.Status = string(pl.Status)
		}
		if len(pl.Tracks) != 0 {
			payload.PeerState.Tracks = make([]*rtc.Event_Track, len(pl.Tracks))
			for i, t := range pl.Tracks {
				payload.PeerState.Tracks[i] = &rtc.Event_Track{
					Id:   t.ID,
					Hint: string(t.Hint),
				}
			}
		}
	}
	return &rtc.Event{
		Id:        id,
		OwnerId:   owner,
		RoomId:    room,
		Payload:   &payload,
		CreatedAt: event.CreatedAt.UnixMilli(),
	}
}

func FromProto(event *rtc.Event) Event {
	var id, owner, room uuid.UUID
	_ = id.UnmarshalBinary(event.Id)
	_ = owner.UnmarshalBinary(event.OwnerId)
	_ = room.UnmarshalBinary(event.RoomId)
	var payload Payload
	if event.Payload != nil && event.Payload.Message != nil {
		pl := event.Payload.Message
		payload.Message = &PayloadMessage{
			Text: pl.Text,
		}
	}
	if event.Payload.PeerState != nil {
		pl := event.Payload.PeerState
		payload.PeerState = &PayloadPeerState{}
		if pl.Status != "" {
			payload.PeerState.Status = PeerStatus(pl.Status)
		}
		if len(pl.Tracks) != 0 {
			payload.PeerState.Tracks = make([]Track, len(pl.Tracks))
			for i, t := range pl.Tracks {
				payload.PeerState.Tracks[i] = Track{
					ID:   t.Id,
					Hint: TrackHint(t.Hint),
				}
			}
		}
	}
	return Event{
		ID:        id,
		Owner:     owner,
		Room:      room,
		Payload:   payload,
		CreatedAt: time.UnixMilli(event.CreatedAt),
	}
}

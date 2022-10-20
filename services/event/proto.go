package event

import (
	"time"

	"github.com/aromancev/confa/internal/proto/rtc"
	"github.com/google/uuid"
)

func ToProto(event Event) *rtc.Event {
	id, _ := event.ID.MarshalBinary()
	room, _ := event.Room.MarshalBinary()
	var payload rtc.Event_Payload
	if event.Payload.PeerState != nil {
		pl := *event.Payload.PeerState
		peer, _ := pl.Peer.MarshalBinary()
		payload.PeerState = &rtc.Event_Payload_PayloadPeerState{
			PeerId: peer,
		}
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
	if event.Payload.Message != nil {
		pl := *event.Payload.Message
		from, _ := pl.From.MarshalBinary()
		payload.Message = &rtc.Event_Payload_PayloadMessage{
			FromId: from,
			Text:   pl.Text,
		}
	}
	if event.Payload.Recording != nil {
		pl := *event.Payload.Recording
		payload.Recording = &rtc.Event_Payload_PayloadRecording{
			Status: string(pl.Status),
		}
	}
	if event.Payload.TrackRecording != nil {
		pl := *event.Payload.TrackRecording
		binID, _ := pl.ID.MarshalBinary()
		payload.TrackRecording = &rtc.Event_Payload_PayloadTrackRecording{
			Id:      binID,
			TrackId: pl.TrackID,
		}
	}
	return &rtc.Event{
		Id:        id,
		RoomId:    room,
		Payload:   &payload,
		CreatedAt: event.CreatedAt.UnixMilli(),
	}
}

func FromProto(event *rtc.Event) Event {
	var id, room uuid.UUID
	_ = id.UnmarshalBinary(event.Id)
	_ = room.UnmarshalBinary(event.RoomId)
	var payload Payload
	if event.Payload.PeerState != nil {
		pl := event.Payload.PeerState
		var peer uuid.UUID
		_ = peer.UnmarshalBinary(pl.PeerId)
		payload.PeerState = &PayloadPeerState{
			Peer: peer,
		}
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
	if event.Payload.Message != nil {
		pl := event.Payload.Message
		var from uuid.UUID
		_ = from.UnmarshalBinary(pl.FromId)
		payload.Message = &PayloadMessage{
			From: from,
			Text: pl.Text,
		}
	}
	if event.Payload.Recording != nil {
		pl := event.Payload.Recording
		payload.Recording = &PayloadRecording{
			Status: RecordStatus(pl.Status),
		}
	}
	if event.Payload.TrackRecording != nil {
		pl := event.Payload.TrackRecording
		var id uuid.UUID
		_ = id.UnmarshalBinary(pl.Id)
		payload.TrackRecording = &PayloadTrackRecording{
			ID:      id,
			TrackID: pl.TrackId,
		}
	}
	return Event{
		ID:        id,
		Room:      room,
		Payload:   payload,
		CreatedAt: time.UnixMilli(event.CreatedAt),
	}
}

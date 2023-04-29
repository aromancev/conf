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
		session, _ := pl.SessionID.MarshalBinary()
		peerState := rtc.Event_Payload_PeerState{
			PeerId:    peer,
			SessionId: session,
		}
		if pl.Status != "" {
			peerState.Status = string(pl.Status)
		}
		payload.Payload = &rtc.Event_Payload_PeerState_{
			PeerState: &peerState,
		}
	}
	if event.Payload.Message != nil {
		pl := *event.Payload.Message
		from, _ := pl.From.MarshalBinary()
		payload.Payload = &rtc.Event_Payload_Message_{
			Message: &rtc.Event_Payload_Message{
				FromId: from,
				Text:   pl.Text,
			},
		}
	}
	if event.Payload.Recording != nil {
		pl := *event.Payload.Recording
		payload.Payload = &rtc.Event_Payload_Recording_{
			Recording: &rtc.Event_Payload_Recording{
				Status: string(pl.Status),
			},
		}
	}
	if event.Payload.TrackRecord != nil {
		pl := *event.Payload.TrackRecord
		binID, _ := pl.RecordID.MarshalBinary()
		record := &rtc.Event_Payload_TrackRecord{
			RecordId: binID,
		}
		switch pl.Kind {
		case TrackKindAudio:
			record.Kind = rtc.TrackKind_AUDIO
		case TrackKindVideo:
			record.Kind = rtc.TrackKind_VIDEO
		}
		switch pl.Source {
		case TrackSourceCamera:
			record.Source = rtc.TrackSource_CAMERA
		case TrackSourceMicrophone:
			record.Source = rtc.TrackSource_MICROPHONE
		case TrackSourceScreen:
			record.Source = rtc.TrackSource_SCREEN
		case TrackSourceScreenAudio:
			record.Source = rtc.TrackSource_SCREEN_AUDIO
		default:
			record.Source = rtc.TrackSource_UNKNOWN
		}
		payload.Payload = &rtc.Event_Payload_TrackRecord_{
			TrackRecord: record,
		}
	}
	if event.Payload.Reaction != nil {
		pl := *event.Payload.Reaction
		from, _ := pl.From.MarshalBinary()
		reaction := rtc.Event_Payload_Reaction{
			FromId: from,
		}
		if pl.Reaction.Clap != nil {
			react := pl.Reaction.Clap
			reaction.Reaction = &rtc.Event_Payload_Reaction_Reaction{
				Reaction: &rtc.Event_Payload_Reaction_Reaction_Clap_{
					Clap: &rtc.Event_Payload_Reaction_Reaction_Clap{
						IsStarting: react.IsStarting,
					},
				},
			}
		}
		payload.Payload = &rtc.Event_Payload_Reaction_{
			Reaction: &reaction,
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
	switch pl := event.Payload.Payload.(type) {
	case *rtc.Event_Payload_PeerState_:
		var peer, session uuid.UUID
		_ = peer.UnmarshalBinary(pl.PeerState.PeerId)
		_ = session.UnmarshalBinary(pl.PeerState.SessionId)
		payload.PeerState = &PayloadPeerState{
			Peer:      peer,
			SessionID: session,
		}
		if pl.PeerState.Status != "" {
			payload.PeerState.Status = PeerStatus(pl.PeerState.Status)
		}
	case *rtc.Event_Payload_Message_:
		var from uuid.UUID
		_ = from.UnmarshalBinary(pl.Message.FromId)
		payload.Message = &PayloadMessage{
			From: from,
			Text: pl.Message.Text,
		}
	case *rtc.Event_Payload_Recording_:
		payload.Recording = &PayloadRecording{
			Status: RecordStatus(pl.Recording.Status),
		}
	case *rtc.Event_Payload_TrackRecord_:
		var id uuid.UUID
		_ = id.UnmarshalBinary(pl.TrackRecord.RecordId)
		record := &PayloadTrackRecord{
			RecordID: id,
		}
		switch pl.TrackRecord.Kind {
		case rtc.TrackKind_AUDIO:
			record.Kind = TrackKindAudio
		case rtc.TrackKind_VIDEO:
			record.Kind = TrackKindVideo
		}
		switch pl.TrackRecord.Source {
		case rtc.TrackSource_CAMERA:
			record.Source = TrackSourceCamera
		case rtc.TrackSource_MICROPHONE:
			record.Source = TrackSourceMicrophone
		case rtc.TrackSource_SCREEN:
			record.Source = TrackSourceScreen
		case rtc.TrackSource_SCREEN_AUDIO:
			record.Source = TrackSourceScreenAudio
		default:
			record.Source = TrackSourceUnknown
		}
		payload.TrackRecord = record
	case *rtc.Event_Payload_Reaction_:
		var from uuid.UUID
		_ = from.UnmarshalBinary(pl.Reaction.FromId)
		payload.Reaction = &PayloadReaction{
			From: from,
		}
		if reaction, ok := pl.Reaction.Reaction.Reaction.(*rtc.Event_Payload_Reaction_Reaction_Clap_); ok {
			payload.Reaction.Reaction.Clap = &ReactionClap{
				IsStarting: reaction.Clap.IsStarting,
			}
		}
	}
	return Event{
		ID:        id,
		Room:      room,
		Payload:   payload,
		CreatedAt: time.UnixMilli(event.CreatedAt),
	}
}

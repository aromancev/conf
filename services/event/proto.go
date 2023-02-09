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
		if len(pl.Tracks) != 0 {
			peerState.Tracks = make([]*rtc.Event_Track, len(pl.Tracks))
			for i, t := range pl.Tracks {
				peerState.Tracks[i] = &rtc.Event_Track{
					Id:   t.ID,
					Hint: string(t.Hint),
				}
			}
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
	if event.Payload.TrackRecording != nil {
		pl := *event.Payload.TrackRecording
		binID, _ := pl.ID.MarshalBinary()
		payload.Payload = &rtc.Event_Payload_TrackRecording_{
			TrackRecording: &rtc.Event_Payload_TrackRecording{
				Id:      binID,
				TrackId: pl.TrackID,
			},
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
		if len(pl.PeerState.Tracks) != 0 {
			payload.PeerState.Tracks = make([]Track, len(pl.PeerState.Tracks))
			for i, t := range pl.PeerState.Tracks {
				payload.PeerState.Tracks[i] = Track{
					ID:   t.Id,
					Hint: TrackHint(t.Hint),
				}
			}
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
	case *rtc.Event_Payload_TrackRecording_:
		var id uuid.UUID
		_ = id.UnmarshalBinary(pl.TrackRecording.Id)
		payload.TrackRecording = &PayloadTrackRecording{
			ID:      id,
			TrackID: pl.TrackRecording.TrackId,
		}
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

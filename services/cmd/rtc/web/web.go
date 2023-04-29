package web

import (
	"github.com/aromancev/confa/event"
)

func NewRoomEvent(ev event.Event) *RoomEvent {
	sev := RoomEvent{
		ID:        ev.ID.String(),
		RoomID:    ev.Room.String(),
		CreatedAt: float64(ev.CreatedAt.UTC().UnixMilli()),
	}
	var payload EventPayload
	switch {
	case ev.Payload.PeerState != nil:
		pl := *ev.Payload.PeerState
		state := EventPeerState{
			PeerID:    pl.Peer.String(),
			SessionID: pl.SessionID.String(),
		}
		if pl.Status != "" {
			state.Status = (*PeerStatus)(&pl.Status)
		}
		payload.PeerState = &state
	case ev.Payload.Message != nil:
		pl := *ev.Payload.Message
		payload.Message = &EventMessage{
			FromID: pl.From.String(),
			Text:   pl.Text,
		}
	case ev.Payload.Recording != nil:
		pl := *ev.Payload.Recording
		payload.Recording = &EventRecording{
			Status: RecordingEventStatus(pl.Status),
		}
	case ev.Payload.TrackRecord != nil:
		pl := *ev.Payload.TrackRecord
		payload.TrackRecord = &EventTrackRecord{
			RecordID: pl.RecordID.String(),
			Kind:     TrackKind(pl.Kind),
			Source:   TrackSource(pl.Source),
		}
	case ev.Payload.Reaction != nil:
		pl := *ev.Payload.Reaction
		var reaction Reaction
		if pl.Reaction.Clap != nil {
			react := pl.Reaction.Clap
			reaction.Clap = &ReactionClap{
				IsStarting: react.IsStarting,
			}
		}
		payload.Reaction = &EventReaction{
			FromID:   pl.From.String(),
			Reaction: reaction,
		}
	}
	sev.Payload = payload
	return &sev
}

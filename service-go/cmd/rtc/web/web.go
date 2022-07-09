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
			PeerID: pl.Peer.String(),
		}
		if pl.Status != "" {
			state.Status = (*PeerStatus)(&pl.Status)
		}
		if len(pl.Tracks) != 0 {
			tracks := make([]Track, len(pl.Tracks))
			for i, t := range pl.Tracks {
				tracks[i] = Track{
					ID:   t.ID,
					Hint: Hint(t.Hint),
				}
			}
			state.Tracks = tracks
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
			Status: RecordingStatus(pl.Status),
		}
	case ev.Payload.TrackRecording != nil:
		pl := *ev.Payload.TrackRecording
		payload.TrackRecording = &EventTrackRecording{
			ID:      pl.ID.String(),
			TrackID: pl.TrackID,
		}
	}
	sev.Payload = payload
	return &sev
}

package web

import (
	"github.com/aromancev/confa/event"
)

func NewRoomEvent(ev event.Event) *RoomEvent {
	sev := RoomEvent{
		ID:        ev.ID.String(),
		OwnerID:   ev.Owner.String(),
		RoomID:    ev.Room.String(),
		CreatedAt: float64(ev.CreatedAt.UTC().UnixMilli()),
	}
	var payload EventPayload
	switch {
	case ev.Payload.PeerState != nil:
		pl := *ev.Payload.PeerState
		var state EventPeerState
		if pl.Status != "" {
			state.Status = (*Status)(&pl.Status)
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
			Text: pl.Text,
		}
	}
	sev.Payload = payload
	return &sev
}

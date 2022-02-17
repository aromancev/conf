package rpc

import (
	"context"
	"fmt"

	"github.com/aromancev/confa/internal/proto/rtc"
	"github.com/aromancev/confa/room"
	"github.com/google/uuid"
)

type Handler struct {
	rooms *room.Mongo
}

func NewHandler(rooms *room.Mongo) *Handler {
	return &Handler{
		rooms: rooms,
	}
}

func (h *Handler) CreateRoom(ctx context.Context, request *rtc.Room) (*rtc.Room, error) {
	var ownerID uuid.UUID
	err := ownerID.UnmarshalBinary(request.OwnerId)
	if err != nil {
		return nil, fmt.Errorf("invalid owner id:%w", err)
	}
	created, err := h.rooms.Create(ctx, room.Room{
		ID:    uuid.New(),
		Owner: ownerID,
	})
	if err != nil {
		return nil, err
	}
	roomID, _ := created[0].ID.MarshalBinary()
	return &rtc.Room{
		Id: roomID,
	}, nil
}

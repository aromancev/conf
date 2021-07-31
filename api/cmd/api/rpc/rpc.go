package rpc

import (
	"context"
	"fmt"

	"github.com/aromancev/confa/internal/room"
	"github.com/aromancev/confa/proto/rtc"
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
	ownerID, err := uuid.Parse(request.OwnerId)
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
	return &rtc.Room{
		Id: created[0].ID.String(),
	}, nil
}

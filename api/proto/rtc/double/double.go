package double

import (
	"context"
	"errors"

	"github.com/aromancev/confa/proto/rtc"
	"github.com/google/uuid"
)

type Memory struct {
	rooms map[uuid.UUID]*rtc.Room
}

func NewMemory() *Memory {
	return &Memory{
		rooms: map[uuid.UUID]*rtc.Room{},
	}
}

func (m *Memory) CreateRoom(ctx context.Context, request *rtc.Room) (*rtc.Room, error) {
	id := uuid.New()
	m.rooms[id] = request
	request.Id = id.String()
	return request, nil
}

func (m *Memory) Room(ctx context.Context, roomID string) (*rtc.Room, error) {
	id, err := uuid.Parse(roomID)
	if err != nil {
		return nil, err
	}
	room, ok := m.rooms[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return room, nil
}

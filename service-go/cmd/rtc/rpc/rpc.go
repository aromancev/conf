package rpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/internal/proto/rtc"
	"github.com/aromancev/confa/internal/proto/tracker"
	"github.com/aromancev/confa/room"
	"github.com/aromancev/confa/room/record"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/twitchtv/twirp"
)

type Tracker interface {
	Start(context.Context, *tracker.StartParams) (*tracker.Tracker, error)
	Stop(context.Context, *tracker.StopParams) (*tracker.Tracker, error)
}

type EventEmitter interface {
	EmitEvent(ctx context.Context, event event.Event) error
}

type Handler struct {
	rooms   *room.Mongo
	records *record.Mongo
	tracker Tracker
	emitter EventEmitter
}

func NewHandler(rooms *room.Mongo, records *record.Mongo, tr Tracker, emitter EventEmitter) *Handler {
	return &Handler{
		rooms:   rooms,
		records: records,
		tracker: tr,
		emitter: emitter,
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
		Id:      roomID,
		OwnerId: request.OwnerId,
	}, nil
}

func (h *Handler) StartRecording(ctx context.Context, request *rtc.RecordingParams) (*rtc.Recording, error) {
	var roomID uuid.UUID
	err := roomID.UnmarshalBinary(request.RoomId)
	if err != nil {
		return nil, fmt.Errorf("invalid owner id:%w", err)
	}

	_, err = h.rooms.FetchOne(ctx, room.Lookup{
		ID: roomID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch room: %w", err)
	}

	newRecord := record.Record{
		ID:   uuid.New(),
		Room: roomID,
		Key:  request.Key,
	}
	upsertedRecord, err := h.records.FetchOrStart(ctx, newRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to start recording: %w", err)
	}
	alreadyExists := newRecord.ID != upsertedRecord.ID

	_, err = h.tracker.Start(ctx, &tracker.StartParams{
		RoomId:     request.RoomId,
		Role:       tracker.Role_RECORD,
		ExpireInMs: request.ExpireInMs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start tracker: %w", err)
	}

	if !alreadyExists {
		err := h.emitter.EmitEvent(ctx, event.Event{
			ID:   uuid.New(),
			Room: roomID,
			Payload: event.Payload{
				Recording: &event.PayloadRecording{
					Status: event.RecordingStarted,
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to emit recording event: %w", err)
		}
	}

	recordIDBin, _ := upsertedRecord.ID.MarshalBinary()

	log.Ctx(ctx).Info().Str("roomId", roomID.String()).Msg("Recording started.")
	return &rtc.Recording{
		RoomId:        request.RoomId,
		RecordingId:   recordIDBin,
		StartedAt:     upsertedRecord.StartedAt.UnixMilli(),
		AlreadyExists: alreadyExists,
	}, nil
}

func (h *Handler) StopRecording(ctx context.Context, request *rtc.RecordingLookup) (*rtc.Recording, error) {
	var roomID uuid.UUID
	err := roomID.UnmarshalBinary(request.RoomId)
	if err != nil {
		return nil, fmt.Errorf("invalid room id: %w", err)
	}

	lookup := record.Lookup{
		Room: roomID,
	}
	if len(request.RecordingId) != 0 {
		var recID uuid.UUID
		err := recID.UnmarshalBinary(request.RecordingId)
		if err != nil {
			return nil, fmt.Errorf("invalid recording id: %w", err)
		}
		lookup.ID = recID
	} else {
		lookup.Key = request.Key
	}

	res, err := h.records.Stop(ctx, lookup)
	switch {
	case err != nil:
		return nil, fmt.Errorf("failed to stop record: %w", err)
	case res.ModifiedCount == 0:
		return nil, twirp.NewError(twirp.NotFound, "Record not found or already stopped.")
	}

	_, err = h.tracker.Stop(ctx, &tracker.StopParams{
		RoomId: request.RoomId,
		Role:   tracker.Role_RECORD,
	})
	var twerr twirp.Error
	switch {
	case errors.As(err, &twerr) && twerr.Code() == twirp.NotFound:
		return nil, twirp.NewError(twirp.NotFound, "Tracker not found or already stopped.")
	case err != nil:
		return nil, fmt.Errorf("failed to stop tracker: %w", err)
	}

	err = h.emitter.EmitEvent(ctx, event.Event{
		ID:   uuid.New(),
		Room: roomID,
		Payload: event.Payload{
			Recording: &event.PayloadRecording{
				Status: event.RecordingStopped,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to emit recording event: %w", err)
	}

	log.Ctx(ctx).Info().Str("roomId", roomID.String()).Msg("Recording stopped.")
	return &rtc.Recording{
		RoomId: request.RoomId,
	}, nil
}

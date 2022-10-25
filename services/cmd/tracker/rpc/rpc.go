package rpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	evtrack "github.com/aromancev/confa/event/tracker"
	pb "github.com/aromancev/confa/internal/proto/tracker"
	"github.com/aromancev/confa/internal/tracker"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	sdk "github.com/pion/ion-sdk-go"
	"github.com/twitchtv/twirp"
)

type Buckets struct {
	TrackRecords string
}

type Handler struct {
	connector    *sdk.Connector
	runtime      *tracker.Runtime
	storage      *minio.Client
	trackEmitter evtrack.TrackEmitter
	eventEmitter evtrack.EventEmitter
	buckets      Buckets
}

func NewHandler(connector *sdk.Connector, runtime *tracker.Runtime, storage *minio.Client, trackEmitter evtrack.TrackEmitter, eventEmitter evtrack.EventEmitter, buckets Buckets) *Handler {
	return &Handler{
		connector:    connector,
		runtime:      runtime,
		storage:      storage,
		buckets:      buckets,
		trackEmitter: trackEmitter,
		eventEmitter: eventEmitter,
	}
}

func (h *Handler) Start(ctx context.Context, params *pb.StartParams) (*pb.Tracker, error) {
	var roomID uuid.UUID
	err := roomID.UnmarshalBinary(params.RoomId)
	if err != nil {
		return nil, fmt.Errorf("faield to unmarshal room id: %w", err)
	}

	tr, err := h.runtime.StartTracker(
		ctx,
		roomID,
		params.Role.String(),
		time.Now().Add(time.Duration(params.ExpireInMs)*time.Millisecond),
		func(ctx context.Context, roomID uuid.UUID) (tracker.Tracker, error) {
			return evtrack.NewTracker(ctx, h.storage, h.connector, h.trackEmitter, h.eventEmitter, h.buckets.TrackRecords, roomID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("faield to start tracker: %w", err)
	}
	return &pb.Tracker{
		RoomId:        params.RoomId,
		AlreadyExists: tr.AlreadyExists,
		StartedAt:     tr.StartedAt.UnixMilli(),
		ExpiresAt:     tr.ExpiresAt.UnixMilli(),
	}, nil
}

func (h *Handler) Stop(ctx context.Context, params *pb.StopParams) (*pb.Tracker, error) {
	var roomID uuid.UUID
	err := roomID.UnmarshalBinary(params.RoomId)
	if err != nil {
		return nil, fmt.Errorf("faield to unmarshal room id: %w", err)
	}

	tr, err := h.runtime.StopTracker(
		ctx,
		roomID,
		params.Role.String(),
	)
	switch {
	case errors.Is(err, tracker.ErrNotFound):
		return nil, twirp.NewError(twirp.NotFound, "Tracker not found.")
	case err != nil:
		return nil, fmt.Errorf("faield to stop tracker: %w", err)
	}
	return &pb.Tracker{
		RoomId:        params.RoomId,
		AlreadyExists: tr.AlreadyExists,
		StartedAt:     tr.StartedAt.UnixMilli(),
		ExpiresAt:     tr.ExpiresAt.UnixMilli(),
	}, nil
}

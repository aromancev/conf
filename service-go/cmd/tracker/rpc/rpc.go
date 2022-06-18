package rpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/aromancev/confa/internal/proto/tracker"
	"github.com/aromancev/confa/tracker/record"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	sdk "github.com/pion/ion-sdk-go"

	"github.com/aromancev/confa/tracker"
)

type Buckets struct {
	TrackRecords string
}

type Handler struct {
	connector    *sdk.Connector
	runtime      *tracker.Runtime
	storage      *minio.Client
	trackEmitter record.Emitter
	buckets      Buckets
}

func NewHandler(connector *sdk.Connector, runtime *tracker.Runtime, storage *minio.Client, trackEmitter record.Emitter, buckets Buckets) *Handler {
	return &Handler{
		connector:    connector,
		runtime:      runtime,
		storage:      storage,
		buckets:      buckets,
		trackEmitter: trackEmitter,
	}
}

func (h *Handler) Join(ctx context.Context, params *pb.JoinParams) (*pb.Joined, error) {
	var roomID uuid.UUID
	err := roomID.UnmarshalBinary(params.RoomId)
	if err != nil {
		return nil, fmt.Errorf("faield to unmarshal room id: %w", err)
	}

	_, err = h.runtime.StartTracker(ctx, roomID, "role", time.Now().Add(time.Hour), func(ctx context.Context, roomID uuid.UUID) (tracker.Tracker, error) {
		return record.NewTracker(ctx, h.storage, h.connector, h.trackEmitter, h.buckets.TrackRecords, roomID)
	})
	if err != nil {
		return nil, fmt.Errorf("faield to start record tracker: %w", err)
	}
	return &pb.Joined{}, nil
}

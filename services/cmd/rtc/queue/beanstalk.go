package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/confa"
	"github.com/aromancev/confa/internal/proto/queue"
	"github.com/google/uuid"
	"github.com/prep/beanstalk"
	"google.golang.org/protobuf/proto"
)

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Beanstalk struct {
	producer Producer
	tubes    Tubes
}

func NewBeanstalk(producer Producer, tubes Tubes) *Beanstalk {
	return &Beanstalk{
		producer: producer,
		tubes:    tubes,
	}
}

func (b *Beanstalk) RecordingReady(ctx context.Context, roomID, recordingID uuid.UUID) error {
	room, _ := roomID.MarshalBinary()
	recording, _ := recordingID.MarshalBinary()
	update := &confa.RecordingUpdate{
		RoomId:      room,
		RecordingId: recording,
		UpdatedAt:   time.Now().UnixMilli(),
		Update: &confa.RecordingUpdate_Update{
			Update: &confa.RecordingUpdate_Update_ProcessingFinished{
				ProcessingFinished: &confa.RecordingUpdate_ProcessingFinished{},
			},
		},
	}
	payload, err := proto.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal recording update: %w", err)
	}
	body, err := proto.Marshal(
		&queue.Job{
			Payload: payload,
			TraceId: trace.ID(ctx),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	_, err = b.producer.Put(ctx, b.tubes.RecordingUpdate, body, beanstalk.PutParams{TTR: 30 * time.Second})
	if err != nil {
		return fmt.Errorf("failed to put job: %w", err)
	}
	return nil
}

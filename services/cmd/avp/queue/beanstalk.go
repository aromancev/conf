package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/queue"
	"github.com/aromancev/confa/internal/proto/rtc"
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

func (b *Beanstalk) ProcessingStarted(ctx context.Context, recordingID, recordID uuid.UUID) error {
	return b.put(ctx, recordingID, recordID, &rtc.UpdateRecordingTrack_Update{
		Update: &rtc.UpdateRecordingTrack_Update_ProcessingStarted{},
	})
}

func (b *Beanstalk) ProcessingFinished(ctx context.Context, recordingID, recordID uuid.UUID) error {
	return b.put(ctx, recordingID, recordID, &rtc.UpdateRecordingTrack_Update{
		Update: &rtc.UpdateRecordingTrack_Update_ProcessingFinished{},
	})
}

func (b *Beanstalk) put(ctx context.Context, recordingID, recordID uuid.UUID, update *rtc.UpdateRecordingTrack_Update) error {
	recordingIDBin, _ := recordingID.MarshalBinary()
	recordIDBin, _ := recordID.MarshalBinary()

	payload, err := proto.Marshal(&rtc.UpdateRecordingTrack{
		RecordingId: recordingIDBin,
		RecordId:    recordIDBin,
		UpdatedAt:   time.Now().UnixMilli(),
		Update:      update,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal recorting track update: %w", err)
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
	_, err = b.producer.Put(ctx, b.tubes.UpdateRecordingTrack, body, beanstalk.PutParams{TTR: 10 * time.Second})
	if err != nil {
		return fmt.Errorf("failed to put job: %w", err)
	}
	return nil
}

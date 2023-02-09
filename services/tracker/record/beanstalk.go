package record

import (
	"context"
	"fmt"
	"time"

	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/avp"
	"github.com/aromancev/confa/internal/proto/queue"
	"github.com/aromancev/confa/internal/proto/rtc"
	"github.com/google/uuid"
	"github.com/prep/beanstalk"
	"google.golang.org/protobuf/proto"
)

type BeanstalkProducer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Tubes struct {
	ProcessTrack         string
	StoreEvent           string
	UpdateRecordingTrack string
}

type Beanstalk struct {
	producer BeanstalkProducer
	tubes    Tubes
}

func NewBeanstalk(producer BeanstalkProducer, tubes Tubes) *Beanstalk {
	return &Beanstalk{
		producer: producer,
		tubes:    tubes,
	}
}

func (b *Beanstalk) RecordStarted(ctx context.Context, record Record) error {
	id, _ := uuid.New().MarshalBinary()
	roomID, _ := record.RoomID.MarshalBinary()
	recordingID, _ := record.RecordingID.MarshalBinary()
	recordID, _ := record.RecordID.MarshalBinary()

	payload, err := proto.Marshal(&rtc.StoreEvent{
		Event: &rtc.Event{
			Id:     id,
			RoomId: roomID,
			Payload: &rtc.Event_Payload{
				Payload: &rtc.Event_Payload_TrackRecording_{
					TrackRecording: &rtc.Event_Payload_TrackRecording{
						Id:      recordID,
						TrackId: record.TrackID,
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
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
	_, err = b.producer.Put(ctx, b.tubes.StoreEvent, body, beanstalk.PutParams{TTR: 10 * time.Second})
	if err != nil {
		return fmt.Errorf("failed to put job: %w", err)
	}

	payload, err = proto.Marshal(&rtc.UpdateRecordingTrack{
		RecordingId: recordingID,
		RecordId:    recordID,
		UpdatedAt:   time.Now().UnixMilli(),
		Update: &rtc.UpdateRecordingTrack_Update{
			Update: &rtc.UpdateRecordingTrack_Update_RecordingStarted{},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal recorting track update: %w", err)
	}
	body, err = proto.Marshal(
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

func (b *Beanstalk) RecordFinished(ctx context.Context, record Record) error {
	roomID, _ := record.RoomID.MarshalBinary()
	recordingID, _ := record.RecordingID.MarshalBinary()
	recordID, _ := record.RecordID.MarshalBinary()

	processTrack := avp.ProcessTrack{
		Bucket:          record.Bucket,
		Object:          record.Object,
		RoomId:          roomID,
		RecordingId:     recordingID,
		RecordId:        recordID,
		DurationSeconds: float32(record.Duration.Seconds()),
	}
	if record.Kind == KindAudio {
		processTrack.Kind = avp.ProcessTrack_AUDIO
	} else {
		processTrack.Kind = avp.ProcessTrack_VIDEO
	}
	payload, err := proto.Marshal(&processTrack)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
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
	_, err = b.producer.Put(ctx, b.tubes.ProcessTrack, body, beanstalk.PutParams{
		TTR: 5 * time.Minute, // The job handler should touch it periodically to prevent rescheduling.
	})
	if err != nil {
		return fmt.Errorf("failed to put job: %w", err)
	}

	payload, err = proto.Marshal(&rtc.UpdateRecordingTrack{
		RecordingId: recordingID,
		RecordId:    recordID,
		UpdatedAt:   time.Now().UnixMilli(),
		Update: &rtc.UpdateRecordingTrack_Update{
			Update: &rtc.UpdateRecordingTrack_Update_RecordingFinished{},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal recorting track update: %w", err)
	}
	body, err = proto.Marshal(
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

	return err
}

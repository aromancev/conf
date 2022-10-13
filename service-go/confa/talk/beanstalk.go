package talk

import (
	"context"
	"fmt"
	"time"

	"github.com/aromancev/proto/confa"
	"github.com/aromancev/proto/queue"
	"github.com/aromancev/telemetry"
	"github.com/google/uuid"
	"github.com/prep/beanstalk"
	"google.golang.org/protobuf/proto"
)

type BeanstalkProducer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Tubes struct {
	StartRecording string
	StopRecording  string
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

func (b *Beanstalk) StartRecording(ctx context.Context, talkID, roomID uuid.UUID) error {
	talkIDBin, _ := talkID.MarshalBinary()
	roomIDBin, _ := roomID.MarshalBinary()

	job := confa.StartRecording{
		TalkId: talkIDBin,
		RoomId: roomIDBin,
	}
	payload, err := proto.Marshal(&job)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	body, err := proto.Marshal(
		&queue.Job{
			Payload: payload,
			TraceId: telemetry.ID(ctx),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	_, err = b.producer.Put(ctx, b.tubes.StartRecording, body, beanstalk.PutParams{
		TTR: 30 * time.Second,
	})
	return err
}

func (b *Beanstalk) StopRecording(ctx context.Context, talkID, roomID uuid.UUID, after time.Duration) error {
	talkIDBin, _ := talkID.MarshalBinary()
	roomIDBin, _ := roomID.MarshalBinary()

	job := confa.StopRecording{
		TalkId: talkIDBin,
		RoomId: roomIDBin,
	}
	payload, err := proto.Marshal(&job)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	body, err := proto.Marshal(
		&queue.Job{
			Payload: payload,
			TraceId: telemetry.ID(ctx),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	_, err = b.producer.Put(ctx, b.tubes.StopRecording, body, beanstalk.PutParams{
		Delay: after,
		TTR:   30 * time.Second,
	})
	return err
}

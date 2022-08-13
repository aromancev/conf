package event

import (
	"context"
	"fmt"
	"time"

	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/queue"
	"github.com/aromancev/confa/internal/proto/rtc"
	"github.com/prep/beanstalk"
	"google.golang.org/protobuf/proto"
)

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type BeansltalkEmitter struct {
	producer Producer
	tube     string
}

func NewBeanstalkEmitter(producer Producer, tube string) *BeansltalkEmitter {
	return &BeansltalkEmitter{
		producer: producer,
		tube:     tube,
	}
}

func (e *BeansltalkEmitter) EmitEvent(ctx context.Context, event Event) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("%w: %s", ErrValidation, err)
	}

	payload, err := proto.Marshal(&rtc.StoreEvent{
		Event: ToProto(event),
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
	_, err = e.producer.Put(ctx, e.tube, body, beanstalk.PutParams{TTR: 10 * time.Second})
	if err != nil {
		return fmt.Errorf("failed to put job: %w", err)
	}
	return nil
}

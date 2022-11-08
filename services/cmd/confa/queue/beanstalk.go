package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/queue"
	"github.com/aromancev/confa/internal/proto/sender"
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

func (b *Beanstalk) Send(ctx context.Context, userID uuid.UUID, message *sender.Message) error {
	user, _ := userID.MarshalBinary()
	payload, err := proto.Marshal(&sender.Send{
		Delivery: &sender.Delivery{
			Delivery: &sender.Delivery_Auto_{
				Auto: &sender.Delivery_Auto{
					UserId: user,
				},
			},
		},
		Message: message,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal message to send: %w", err)
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
	_, err = b.producer.Put(ctx, b.tubes.Send, body, beanstalk.PutParams{TTR: 30 * time.Second})
	if err != nil {
		return fmt.Errorf("failed to put job: %w", err)
	}
	return nil
}

package profile

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

type BeanstalkTubes struct {
	UpdateAvatar string
}

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type BeansltalkEmitter struct {
	producer Producer
	tubes    BeanstalkTubes
}

func NewBeanstalkEmitter(producer Producer, tubes BeanstalkTubes) *BeansltalkEmitter {
	return &BeansltalkEmitter{
		producer: producer,
		tubes:    tubes,
	}
}

func (e *BeansltalkEmitter) UpdateProfile(ctx context.Context, userID uuid.UUID, source AvatarSource) error {
	if err := source.Validate(); err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	id, _ := userID.MarshalBinary()
	job := confa.UpdateProfile{
		UserId: id,
		Avatar: &confa.UpdateProfile_Source{},
	}
	if source.PublicURL != nil {
		job.Avatar.PublicUrl = &confa.UpdateProfile_Source_PublicURL{
			Url: source.PublicURL.URL,
		}
	}
	if source.Storage != nil {
		job.Avatar.Storage = &confa.UpdateProfile_Source_Storage{
			Bucket: source.Storage.Bucket,
			Path:   source.Storage.Path,
		}
	}
	payload, err := proto.Marshal(&job)
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
	_, err = e.producer.Put(ctx, e.tubes.UpdateAvatar, body, beanstalk.PutParams{
		Delay: 2 * time.Second,
		TTR:   2 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("failed to put job: %w", err)
	}
	return nil
}

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

func (e *BeansltalkEmitter) UpdateProfile(ctx context.Context, userID uuid.UUID, givenName, familyName *string, thumbnail, avatar FileSource) error {
	if err := thumbnail.Validate(); err != nil {
		return fmt.Errorf("invalid thumbnail: %w", err)
	}
	if err := avatar.Validate(); err != nil {
		return fmt.Errorf("invalid avatar: %w", err)
	}

	id, _ := userID.MarshalBinary()
	job := confa.UpdateProfile{
		UserId:    id,
		Thumbnail: newFileSource(thumbnail),
		Avatar:    newFileSource(avatar),
	}
	if givenName != nil {
		job.GivenNameSet = true
		job.GivenName = *givenName
	}
	if familyName != nil {
		job.FamilyNameSet = true
		job.FamilyName = *familyName
	}
	if thumbnail.PublicURL != nil {
		job.Thumbnail.PublicUrl = &confa.UpdateProfile_FileSource_PublicURL{
			Url: thumbnail.PublicURL.URL,
		}
	}
	if thumbnail.Storage != nil {
		job.Thumbnail.Storage = &confa.UpdateProfile_FileSource_Storage{
			Bucket: thumbnail.Storage.Bucket,
			Path:   thumbnail.Storage.Path,
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

func newFileSource(source FileSource) *confa.UpdateProfile_FileSource {
	if source.PublicURL != nil {
		return &confa.UpdateProfile_FileSource{
			PublicUrl: &confa.UpdateProfile_FileSource_PublicURL{
				Url: source.PublicURL.URL,
			},
		}
	}
	if source.Storage != nil {
		return &confa.UpdateProfile_FileSource{
			Storage: &confa.UpdateProfile_FileSource_Storage{
				Bucket: source.Storage.Bucket,
				Path:   source.Storage.Path,
			},
		}
	}
	return nil
}

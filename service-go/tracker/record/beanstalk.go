package record

import (
	"context"
	"fmt"
	"time"

	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/avp"
	"github.com/aromancev/confa/internal/proto/queue"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/prep/beanstalk"
	"google.golang.org/protobuf/proto"
)

type BeanstalkProducer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Tubes struct {
	ProcessTrack string
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

func (b *Beanstalk) ProcessTrack(ctx context.Context, roomID, recordID uuid.UUID, bucket, object string, kind webrtc.RTPCodecType, duration time.Duration) error {
	roomIDBin, _ := roomID.MarshalBinary()
	recordIDBin, _ := recordID.MarshalBinary()
	job := avp.ProcessTrack{
		Bucket:          bucket,
		Object:          object,
		RoomId:          roomIDBin,
		RecordId:        recordIDBin,
		DurationSeconds: float32(duration.Seconds()),
	}
	if kind == webrtc.RTPCodecTypeVideo {
		job.Kind = avp.ProcessTrack_VIDEO
	} else {
		job.Kind = avp.ProcessTrack_AUDIO
	}
	payload, err := proto.Marshal(&job)
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
	return err
}

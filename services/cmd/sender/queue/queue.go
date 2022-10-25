package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	"github.com/aromancev/confa/internal/platform/backoff"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/queue"
	pb "github.com/aromancev/confa/internal/proto/sender"
)

type Tubes struct {
	Send string
}

type JobHandle func(ctx context.Context, job *beanstalk.Job)

type Handler struct {
	route func(job *beanstalk.Job) JobHandle
}

type Sender interface {
	Send(context.Context, *pb.Send) error
}

func NewHandler(sender Sender, tubes Tubes) *Handler {
	return &Handler{
		route: func(job *beanstalk.Job) JobHandle {
			switch job.Stats.Tube {
			case tubes.Send:
				return send(sender)
			default:
				return nil
			}
		},
	}
}

func (h *Handler) ServeJob(ctx context.Context, job *beanstalk.Job) {
	l := log.Ctx(ctx).With().Uint64("jobId", job.ID).Str("tube", job.Stats.Tube).Logger()
	ctx = l.WithContext(ctx)

	var qJob queue.Job
	err := proto.Unmarshal(job.Body, &qJob)
	if err != nil {
		log.Ctx(ctx).Error().Str("tube", job.Stats.Tube).Msg("Failed to unmarshal job. Burying.")
		return
	}
	ctx = trace.New(ctx, qJob.TraceId)
	job.Body = qJob.Payload

	log.Ctx(ctx).Info().Msg("Job received.")

	defer func() {
		if err := recover(); err != nil {
			log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Msg("ServeJob panic")
		}
	}()

	handle := h.route(job)
	if handle == nil {
		log.Ctx(ctx).Error().Msg("No handle for job. Burying.")
		return
	}

	handle(ctx, job)
}

func send(sender Sender) JobHandle {
	const maxAge = 24 * time.Hour
	bo := backoff.Backoff{
		Factor: 1.2,
		Min:    1 * time.Second,
		Max:    20 * time.Minute,
	}

	return func(ctx context.Context, job *beanstalk.Job) {
		var sendJob pb.Send
		err := proto.Unmarshal(job.Body, &sendJob)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal email job.")
			jobDelete(ctx, job)
			return
		}

		err = sender.Send(ctx, &sendJob)
		if err != nil {
			jobRetry(ctx, job, bo, maxAge)
			return
		}
		jobDelete(ctx, job)
		log.Ctx(ctx).Info().Msg("Message sent.")
	}
}

func jobRetry(ctx context.Context, job *beanstalk.Job, bo backoff.Backoff, maxAge time.Duration) {
	if job.Stats.Age > maxAge {
		log.Ctx(ctx).Error().Int("retries", job.Stats.Releases).Dur("age", job.Stats.Age).Msg("Job retries exceeded. Burying.")
		if err := job.Bury(ctx); err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to bury job")
		}
		return
	}

	if err := job.ReleaseWithParams(ctx, job.Stats.Priority, bo.ForAttempt(float64(job.Stats.Releases))); err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to release job")
		return
	}
	log.Ctx(ctx).Debug().Msg("Job released")
}

func jobDelete(ctx context.Context, job *beanstalk.Job) {
	if err := job.Delete(ctx); err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to delete job")
		return
	}
	log.Ctx(ctx).Info().Msg("Job deleted.")
}

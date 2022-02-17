package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/internal/platform/backoff"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/queue"
)

const (
	jobRetries = 10
)

type JobHandle func(ctx context.Context, job *beanstalk.Job) error

type Handler struct {
	route func(job *beanstalk.Job) JobHandle
}

func NewHandler(events *event.Mongo) *Handler {
	return &Handler{
		route: func(job *beanstalk.Job) JobHandle {
			switch job.Stats.Tube {
			case queue.TubeEvent:
				return saveEvent(events)
			default:
				return nil
			}
		},
	}
}

func (h *Handler) ServeJob(ctx context.Context, job *beanstalk.Job) {
	l := log.Ctx(ctx).With().Uint64("jobId", job.ID).Str("tube", job.Stats.Tube).Logger()
	ctx = l.WithContext(ctx)

	j, err := queue.Unmarshal(job.Body)
	if err != nil {
		log.Ctx(ctx).Error().Str("tube", job.Stats.Tube).Msg("No handle for job. Burying.")
		return
	}
	ctx = trace.New(ctx, j.TraceId)
	job.Body = j.Payload

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

	err = handle(ctx, job)
	if err != nil {
		if job.Stats.Releases >= jobRetries {
			log.Ctx(ctx).Err(err).Msg("Job retries exceeded. Burying.")
			if err := job.Bury(ctx); err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to bury job")
			}
			return
		}

		bo := backoff.Backoff{
			Factor: 1.2,
			Min:    10 * time.Second,
			Max:    10 * time.Minute,
		}
		log.Ctx(ctx).Err(err).Msg("Job failed. Releasing.")
		if err := job.ReleaseWithParams(ctx, job.Stats.Priority, bo.ForAttempt(float64(job.Stats.Releases))); err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to release job")
		}
		return
	}

	if err := job.Delete(ctx); err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to delete job")
	}
	log.Ctx(ctx).Info().Msg("Job served.")
}

package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/aromancev/avp/dash"
	"github.com/aromancev/avp/internal/platform/backoff"
	"github.com/aromancev/proto/avp"
	"github.com/aromancev/proto/queue"
	"github.com/aromancev/telemetry"
	"github.com/google/uuid"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type Tubes struct {
	ProcessTrack string
}

type JobHandle func(ctx context.Context, job *beanstalk.Job)

type Handler struct {
	route func(job *beanstalk.Job) JobHandle
}

func NewHandler(converter *dash.Converter, tubes Tubes) *Handler {
	return &Handler{
		route: func(job *beanstalk.Job) JobHandle {
			switch job.Stats.Tube {
			case tubes.ProcessTrack:
				return AutoTouch(processTrack(converter))
			default:
				return nil
			}
		},
	}
}

func (h *Handler) ServeJob(ctx context.Context, job *beanstalk.Job) {
	l := log.Ctx(ctx).With().Uint64("jobId", job.ID).Str("tube", job.Stats.Tube).Logger()
	ctx = l.WithContext(ctx)

	var j queue.Job
	err := proto.Unmarshal(job.Body, &j)
	if err != nil {
		log.Ctx(ctx).Error().Str("tube", job.Stats.Tube).Msg("Failed to unmarshal job. Burying.")
		return
	}
	ctx = telemetry.New(ctx, j.TraceId)
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

	handle(ctx, job)
}

func processTrack(converter *dash.Converter) JobHandle {
	const maxAge = 24 * time.Hour
	bo := backoff.Backoff{
		Factor: 1.5,
		Min:    2 * time.Second,
		Max:    time.Hour,
	}

	return func(ctx context.Context, job *beanstalk.Job) {
		var payload avp.ProcessTrack
		err := proto.Unmarshal(job.Body, &payload)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal event job.")
			jobDelete(ctx, job)
			return
		}

		var roomID, recordID uuid.UUID
		_ = roomID.UnmarshalBinary(payload.RoomId)
		_ = recordID.UnmarshalBinary(payload.RecordId)

		log.Ctx(ctx).Info().Str("roomId", roomID.String()).Str("recordId", recordID.String()).Str("bucket", payload.Bucket).Str("object", payload.Object).Msg("Started processing track.")

		record := dash.Record{
			ID:         recordID,
			BucketName: payload.Bucket,
			ObjectName: payload.Object,
			Duration:   time.Duration(float64(payload.DurationSeconds) * float64(time.Second)),
		}
		if payload.Kind == avp.ProcessTrack_VIDEO {
			err = converter.ConvertVideo(ctx, roomID, record)
		} else {
			err = converter.ConvertAudio(ctx, roomID, record)
		}
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to process track.")
			jobRetry(ctx, job, bo, maxAge)
			return
		}

		log.Ctx(ctx).Info().Msg("Finished processing track.")
		jobDelete(ctx, job)
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

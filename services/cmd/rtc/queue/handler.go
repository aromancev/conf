package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/internal/platform/backoff"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/queue"
	"github.com/aromancev/confa/internal/proto/rtc"
	"github.com/aromancev/confa/room/record"
	"github.com/google/uuid"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type JobHandle func(ctx context.Context, job *beanstalk.Job)

type Emitter interface {
	RecordingReady(ctx context.Context, roomID, recordingID uuid.UUID) error
}

type Handler struct {
	route func(job *beanstalk.Job) JobHandle
}

func NewHandler(events *event.Mongo, records *record.Mongo, tubes Tubes, emitter Emitter) *Handler {
	return &Handler{
		route: func(job *beanstalk.Job) JobHandle {
			switch job.Stats.Tube {
			case tubes.StoreEvent:
				return storeEvent(events)
			case tubes.UpdateRecordingTrack:
				return updateRecordingTrack(records, emitter)
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
		_ = job.Bury(ctx)
		return
	}

	handle(ctx, job)
}

func storeEvent(events *event.Mongo) JobHandle {
	const maxAge = 5 * time.Minute
	bo := backoff.Backoff{
		Factor: 1.5,
		Min:    100 * time.Millisecond,
		Max:    time.Second,
	}

	return func(ctx context.Context, job *beanstalk.Job) {
		var storeEvent rtc.StoreEvent
		err := proto.Unmarshal(job.Body, &storeEvent)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal event job.")
			jobDelete(ctx, job)
			return
		}

		toCreate := event.FromProto(storeEvent.Event)
		ev, err := events.CreateOne(ctx, toCreate)
		switch {
		case errors.Is(err, event.ErrValidation):
			log.Ctx(ctx).Err(err).Msg("Invalid payload for event job. Deleting.")
			jobDelete(ctx, job)
			return
		case errors.Is(err, event.ErrDuplicatedEntry):
			log.Ctx(ctx).Warn().Str("eventId", toCreate.ID.String()).Msg("Skipping duplicated event.")
			jobDelete(ctx, job)
			return
		case err != nil:
			jobRetry(ctx, job, bo, maxAge)
			return
		}
		jobDelete(ctx, job)
		log.Ctx(ctx).Debug().Str("eventId", ev.ID.String()).Msg("Event created.")
	}
}

func updateRecordingTrack(records *record.Mongo, emitter Emitter) JobHandle {
	const maxAge = 5 * time.Minute
	bo := backoff.Backoff{
		Factor: 1.5,
		Min:    100 * time.Millisecond,
		Max:    time.Second,
	}

	return func(ctx context.Context, job *beanstalk.Job) {
		var update rtc.UpdateRecordingTrack
		err := proto.Unmarshal(job.Body, &update)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal udate recording track job.")
			jobDelete(ctx, job)
			return
		}

		var id, recordID uuid.UUID
		_ = id.UnmarshalBinary(update.RecordingId)
		_ = recordID.UnmarshalBinary(update.RecordId)

		var updates record.Records
		switch update.Update.Update.(type) {
		case *rtc.UpdateRecordingTrack_Update_RecordingStarted:
			updates.RecordingStarted = []uuid.UUID{recordID}
		case *rtc.UpdateRecordingTrack_Update_RecordingFinished:
			updates.RecordingFinished = []uuid.UUID{recordID}
		case *rtc.UpdateRecordingTrack_Update_ProcessingStarted:
			updates.ProcessingStarted = []uuid.UUID{recordID}
		case *rtc.UpdateRecordingTrack_Update_ProcessingFinished:
			updates.ProcessingFinished = []uuid.UUID{recordID}
		default:
			log.Ctx(ctx).Error().Msg("Received unknown recording update.")
			jobDelete(ctx, job)
			return
		}
		updated, err := records.UpdateRecords(
			ctx,
			record.Lookup{ID: id},
			updates,
		)
		switch {
		case errors.Is(err, record.ErrValidation):
			log.Ctx(ctx).Err(err).Msg("Failed to update recording.")
			jobDelete(ctx, job)
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to update recording.")
			jobRetry(ctx, job, bo, maxAge)
			return
		}
		if updated.IsReady() {
			err := emitter.RecordingReady(ctx, updated.Room, updated.ID)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to emit recording ready.")
				jobRetry(ctx, job, bo, maxAge)
				return
			}
		}

		jobDelete(ctx, job)
		log.Ctx(ctx).Info().Msg("Recording updated.")
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

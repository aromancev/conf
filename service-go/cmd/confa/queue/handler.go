package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aromancev/confa/internal/platform/backoff"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/confa"
	"github.com/aromancev/confa/internal/proto/queue"
	"github.com/aromancev/confa/profile"
	"github.com/google/uuid"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type Tubes struct {
	UpdateAvatar string
}

type JobHandle func(ctx context.Context, job *beanstalk.Job)

type Handler struct {
	route func(job *beanstalk.Job) JobHandle
}

func NewHandler(uploader *profile.Updater, tubes Tubes) *Handler {
	return &Handler{
		route: func(job *beanstalk.Job) JobHandle {
			switch job.Stats.Tube {
			case tubes.UpdateAvatar:
				return updateAvatar(uploader)
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
		return
	}

	handle(ctx, job)
}

func updateAvatar(uploader *profile.Updater) JobHandle {
	const maxRetries = 10
	bo := backoff.Backoff{
		Factor: 1.5,
		Min:    2 * time.Second,
		Max:    time.Hour,
	}

	return func(ctx context.Context, job *beanstalk.Job) {
		var payload confa.UpdateProfile
		err := proto.Unmarshal(job.Body, &payload)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal event job.")
			jobDelete(ctx, job)
			return
		}
		var userID uuid.UUID
		_ = userID.UnmarshalBinary(payload.UserId)

		var source profile.AvatarSource
		if payload.Avatar.PublicUrl != nil {
			source.PublicURL = &profile.AvatarSourcePublicURL{
				URL: payload.Avatar.PublicUrl.Url,
			}
		}
		if payload.Avatar.Storage != nil {
			source.Storage = &profile.AvatarSourceStorage{
				Bucket: payload.Avatar.Storage.Bucket,
				Path:   payload.Avatar.Storage.Path,
			}
		}

		err = uploader.Update(ctx, userID, source)
		switch {
		case errors.Is(err, profile.ErrValidation):
			log.Ctx(ctx).Err(err).Msg("Invalid payload for update avatar job.")
			jobDelete(ctx, job)
			return
		case errors.Is(err, profile.ErrNotFound):
			log.Ctx(ctx).Debug().Msg("Image is not uploaded yet.")
			jobRetry(ctx, job, bo, maxRetries)
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Unknown error for update avatar job.")
			jobRetry(ctx, job, bo, maxRetries)
			return
		default:
			log.Ctx(ctx).Info().Msg("Avatar update processed.")
			jobDelete(ctx, job)
			return
		}
	}
}

func jobRetry(ctx context.Context, job *beanstalk.Job, bo backoff.Backoff, maxTries int) {
	if job.Stats.Releases >= maxTries {
		log.Ctx(ctx).Error().Int("retries", maxTries).Msg("Job retries exceeded. Burying.")
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

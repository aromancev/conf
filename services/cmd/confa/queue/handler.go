package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aromancev/confa/confa"
	"github.com/aromancev/confa/confa/talk"
	"github.com/aromancev/confa/internal/platform/backoff"
	"github.com/aromancev/confa/internal/platform/trace"
	confapb "github.com/aromancev/confa/internal/proto/confa"
	"github.com/aromancev/confa/internal/proto/queue"
	"github.com/aromancev/confa/internal/proto/rtc"
	senderpb "github.com/aromancev/confa/internal/proto/sender"
	"github.com/aromancev/confa/internal/routes"
	"github.com/aromancev/confa/profile"
	"github.com/google/uuid"
	"github.com/twitchtv/twirp"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type RTC interface {
	StartRecording(context.Context, *rtc.RecordingParams) (*rtc.Recording, error)
	StopRecording(context.Context, *rtc.RecordingLookup) (*rtc.Recording, error)
}

type Sender interface {
	Send(ctx context.Context, userID uuid.UUID, message *senderpb.Message) error
}

type JobHandle func(ctx context.Context, job *beanstalk.Job)

type Handler struct {
	route func(job *beanstalk.Job) JobHandle
}

func NewHandler(uploader *profile.Updater, r RTC, confas *confa.Mongo, talks *talk.Mongo, emitter *talk.Beanstalk, tubes Tubes, sender Sender, rts *routes.Routes) *Handler {
	return &Handler{
		route: func(job *beanstalk.Job) JobHandle {
			switch job.Stats.Tube {
			case tubes.UpdateAvatar:
				return updateAvatar(uploader)
			case tubes.StartRecording:
				return startRecording(talks, r, emitter)
			case tubes.StopRecording:
				return stopRecording(talks, r)
			case tubes.RecordingUpdate:
				return recordingUpdate(confas, talks, sender, rts)
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
		log.Ctx(ctx).Error().Str("tube", job.Stats.Tube).Msg("No handle for job. Burying.")
		_ = job.Bury(ctx)
		return
	}

	handle(ctx, job)
}

func updateAvatar(uploader *profile.Updater) JobHandle {
	const maxAge = 24 * time.Hour
	bo := backoff.Backoff{
		Factor: 1.5,
		Min:    2 * time.Second,
		Max:    time.Hour,
	}

	return func(ctx context.Context, job *beanstalk.Job) {
		var payload confapb.UpdateProfile
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
			jobRetry(ctx, job, bo, maxAge)
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Unknown error for update avatar job.")
			jobRetry(ctx, job, bo, maxAge)
			return
		default:
			log.Ctx(ctx).Info().Msg("Avatar update processed.")
			jobDelete(ctx, job)
			return
		}
	}
}

func startRecording(talks *talk.Mongo, rtcClient RTC, emitter *talk.Beanstalk) JobHandle {
	const maxAge = 2 * time.Minute
	const autostopAfter = 60 * time.Minute
	const maxDuration = time.Hour
	bo := backoff.Backoff{
		Factor: 1.2,
		Min:    1 * time.Second,
		Max:    10 * time.Second,
	}

	return func(ctx context.Context, job *beanstalk.Job) {
		var payload confapb.StartRecording
		err := proto.Unmarshal(job.Body, &payload)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal event job.")
			jobDelete(ctx, job)
			return
		}
		var talkID, roomID uuid.UUID
		_ = talkID.UnmarshalBinary(payload.TalkId)
		_ = roomID.UnmarshalBinary(payload.RoomId)

		recording, err := rtcClient.StartRecording(ctx, &rtc.RecordingParams{
			RoomId:     payload.RoomId,
			Key:        talkID.String(),
			ExpireInMs: maxDuration.Milliseconds(),
		})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to start recording.")
			jobRetry(ctx, job, bo, maxAge)
			return
		}
		if !recording.AlreadyExists {
			err := emitter.StopRecording(ctx, talkID, roomID, autostopAfter)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to emit stop recording.")
				jobRetry(ctx, job, bo, maxAge)
				return
			}
		}

		stateRecording := talk.StateRecording
		_, err = talks.UpdateOne(
			ctx,
			talk.Lookup{
				ID:      talkID,
				StateIn: []talk.State{talk.StateLive},
			},
			talk.Mask{
				State: &(stateRecording),
			},
		)
		switch {
		case errors.Is(err, talk.ErrNotFound):
			log.Ctx(ctx).Warn().Msg("Talk already started.")
			jobDelete(ctx, job)
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to update talk.")
			jobRetry(ctx, job, bo, maxAge)
			return
		}

		log.Ctx(ctx).Info().Msg("Talk recording started.")
		jobDelete(ctx, job)
	}
}

func stopRecording(talks *talk.Mongo, rtcClient RTC) JobHandle {
	const maxAge = 2 * time.Hour
	bo := backoff.Backoff{
		Factor: 1.2,
		Min:    1 * time.Second,
		Max:    10 * time.Minute,
	}

	return func(ctx context.Context, job *beanstalk.Job) {
		var payload confapb.StopRecording
		err := proto.Unmarshal(job.Body, &payload)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal event job.")
			jobDelete(ctx, job)
			return
		}
		var talkID uuid.UUID
		_ = talkID.UnmarshalBinary(payload.TalkId)

		_, err = rtcClient.StopRecording(ctx, &rtc.RecordingLookup{
			RoomId: payload.RoomId,
			Key:    talkID.String(),
		})
		var twerr twirp.Error
		switch {
		// Not found means it's probably already stopped.
		case errors.As(err, &twerr) && twerr.Code() == twirp.NotFound:
			break
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to stop recording.")
			jobRetry(ctx, job, bo, maxAge)
			return
		}

		stateEnded := talk.StateEnded
		_, err = talks.UpdateOne(
			ctx,
			talk.Lookup{
				ID:      talkID,
				StateIn: []talk.State{talk.StateRecording},
			},
			talk.Mask{
				State: &stateEnded,
			},
		)
		switch {
		case errors.Is(err, talk.ErrNotFound):
			log.Ctx(ctx).Info().Msg("Talk already stopped.")
			jobDelete(ctx, job)
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to update talk.")
			jobRetry(ctx, job, bo, maxAge)
			return
		}
		log.Ctx(ctx).Info().Msg("Talk recording stopped.")
		jobDelete(ctx, job)
	}
}

func recordingUpdate(confas *confa.Mongo, talks *talk.Mongo, sender Sender, rts *routes.Routes) JobHandle {
	const maxAge = 2 * time.Hour
	bo := backoff.Backoff{
		Factor: 1.2,
		Min:    1 * time.Second,
		Max:    10 * time.Minute,
	}

	return func(ctx context.Context, job *beanstalk.Job) {
		var payload confapb.RecordingUpdate
		err := proto.Unmarshal(job.Body, &payload)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal event job.")
			jobDelete(ctx, job)
			return
		}
		var roomID uuid.UUID
		_ = roomID.UnmarshalBinary(payload.RoomId)

		if payload.Update == nil {
			log.Ctx(ctx).Error().Msg("Recording update is empty.")
			jobDelete(ctx, job)
			return
		}
		_, ok := payload.Update.Update.(*confapb.RecordingUpdate_Update_ProcessingFinished)
		if !ok {
			log.Ctx(ctx).Info().Msg("Skipping recording update.")
			jobDelete(ctx, job)
			return
		}

		tlk, err := talks.FetchOne(ctx, talk.Lookup{
			Room: roomID,
		})
		switch {
		case errors.Is(err, talk.ErrNotFound):
			log.Ctx(ctx).Err(err).Msg("Failed to find talk for recording.")
			jobDelete(ctx, job)
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to find talk for recording.")
			jobRetry(ctx, job, bo, maxAge)
			return
		}
		cnf, err := confas.FetchOne(ctx, confa.Lookup{
			ID: tlk.Confa,
		})
		switch {
		case errors.Is(err, confa.ErrNotFound):
			log.Ctx(ctx).Err(err).Msg("Failed to find confa for recording.")
			jobDelete(ctx, job)
			return
		case err != nil:
			log.Ctx(ctx).Err(err).Msg("Failed to find confa for recording.")
			jobRetry(ctx, job, bo, maxAge)
			return
		}

		err = sender.Send(ctx, tlk.Owner, &senderpb.Message{
			Message: &senderpb.Message_TalkRecordingReady_{
				TalkRecordingReady: &senderpb.Message_TalkRecordingReady{
					ConfaUrl:   rts.Confa(cnf.Handle),
					ConfaTitle: cnf.Title,
					TalkUrl:    rts.Talk(cnf.Handle, tlk.Handle),
					TalkTitle:  tlk.Title,
				},
			},
		})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to send recording notification.")
			jobRetry(ctx, job, bo, maxAge)
			return
		}

		log.Ctx(ctx).Info().Msg("Talk recording update processed.")
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

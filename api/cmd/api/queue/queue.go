package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	"github.com/aromancev/confa/internal/event"
	"github.com/aromancev/confa/internal/platform/backoff"
	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/proto/queue"
)

const (
	jobRetries = 10
)

type JobHandle func(ctx context.Context, job *beanstalk.Job) error

type Handler struct {
	route func(job *beanstalk.Job) JobHandle
}

func NewHandler(sender *email.Sender, events *event.Mongo) *Handler {
	return &Handler{
		route: func(job *beanstalk.Job) JobHandle {
			switch job.Stats.Tube {
			case queue.TubeEmail:
				return sendEmail(sender)
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

func sendEmail(sender *email.Sender) JobHandle {
	return func(ctx context.Context, j *beanstalk.Job) error {
		var job queue.EmailJob
		err := proto.Unmarshal(j.Body, &job)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal email job.")
			return nil
		}
		emails := make([]email.Email, len(job.Emails))
		for i, e := range job.Emails {
			emails[i] = email.Email{
				FromName:  e.FromName,
				ToAddress: e.ToAddress,
				Subject:   e.Subject,
				HTML:      e.Html,
			}
		}
		err, errs := sender.Send(ctx, emails...)
		if err != nil {
			return err
		}
		for _, err := range errs {
			if err == nil {
				log.Ctx(ctx).Info().Msg("Email sent.")
			} else {
				log.Ctx(ctx).Err(err).Msg("Failed to send email.")
			}
		}
		return nil
	}
}

func saveEvent(events *event.Mongo) JobHandle {
	return func(ctx context.Context, j *beanstalk.Job) error {
		var job queue.EventJob
		err := proto.Unmarshal(j.Body, &job)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal event job.")
			return nil
		}

		var eventID, ownerID, roomID uuid.UUID
		if err := eventID.UnmarshalBinary(job.Id); err != nil {
			return err
		}
		if err := ownerID.UnmarshalBinary(job.OwnerId); err != nil {
			return err
		}
		if err := roomID.UnmarshalBinary(job.RoomId); err != nil {
			return err
		}

		ev := event.Event{
			ID:    eventID,
			Owner: ownerID,
			Room:  roomID,
		}
		switch pl := job.Event.(type) {
		case *queue.EventJob_PeerStatus_:
			ev.Payload = event.Payload{
				Type: event.TypePeerStatus,
				Payload: event.PayloadPeerStatus{
					Status: event.PeerStatus(pl.PeerStatus.Status),
				},
			}
		default:
			return errors.New("unknown event type")
		}

		_, err = events.CreateOne(ctx, ev)
		switch {
		case errors.Is(err, event.ErrValidation):
			log.Ctx(ctx).Err(err).Msg("Invalid payload for event job. Deleting.")
			return nil
		case errors.Is(err, event.ErrDuplicatedEntry):
			log.Ctx(ctx).Warn().Str("eventId", ev.ID.String()).Msg("Skipping duplicated event.")
			return nil
		case err != nil:
			return err
		default:
			return nil
		}
	}
}

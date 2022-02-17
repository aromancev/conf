package queue

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/internal/proto/queue"
	"github.com/google/uuid"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

func saveEvent(events *event.Mongo) JobHandle {
	return func(ctx context.Context, j *beanstalk.Job) error {
		var job queue.EventJob
		err := proto.Unmarshal(j.Body, &job)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal event job.")
			return nil
		}

		var eventID, ownerID, roomID uuid.UUID
		var payload event.Payload
		if err := eventID.UnmarshalBinary(job.Id); err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal event id.")
			return nil
		}
		if err := ownerID.UnmarshalBinary(job.OwnerId); err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal owner id.")
			return nil
		}
		if err := roomID.UnmarshalBinary(job.RoomId); err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal room id.")
			return nil
		}
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal payload.")
			return nil
		}

		ev, err := events.CreateOne(ctx, event.Event{
			ID:      eventID,
			Owner:   ownerID,
			Room:    roomID,
			Payload: payload,
		})
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

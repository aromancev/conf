package handler

import (
	"context"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/proto/queue"
)

func sendEmail(sender *email.Sender) JobHandle {
	return func(ctx context.Context, j *beanstalk.Job) error {
		var job queue.EmailJob
		err := proto.Unmarshal(j.Body, &job)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal email job")
			if err := j.Delete(ctx); err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to delete job")
			}
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
				log.Ctx(ctx).Info().Msg("Email sent")
			} else {
				log.Ctx(ctx).Err(err).Msg("Failed to send email")
			}
		}
		return nil
	}
}

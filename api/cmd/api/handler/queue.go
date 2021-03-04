package handler

import (
	"context"
	"encoding/json"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/internal/platform/email"
)

func (h *Handler) sendEmail(ctx context.Context, job *beanstalk.Job) error {
	var emails []email.Email
	err := json.Unmarshal(job.Body, &emails)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to unmarshal email job")
		if err := job.Delete(ctx); err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to delete job")
		}
		return nil
	}
	err, errs := h.sender.Send(ctx, emails...)
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

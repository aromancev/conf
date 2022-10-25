package sender

import (
	"context"
	"errors"
	"fmt"

	"github.com/aromancev/confa/internal/proto/sender"
)

type EmailSender interface {
	Send(ctx context.Context, message *sender.Message, toAddress string) error
}

type Sender struct {
	email EmailSender
}

func NewSender(email EmailSender) *Sender {
	return &Sender{email: email}
}

func (s *Sender) Send(ctx context.Context, send *sender.Send) error {
	switch delivery := send.Delivery.Delivery.(type) { // nolint: gocritic
	case *sender.Delivery_Email_:
		err := s.email.Send(ctx, send.Message, delivery.Email.ToAddress)
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
		return nil
	}
	return errors.New("unknown delivery method")
}

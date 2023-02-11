package email

import (
	"context"
	"errors"
	"fmt"

	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/proto/sender"
)

type Mailer interface {
	Send(ctx context.Context, emails ...email.Email) error
}

type Sender struct {
	mailer Mailer
	from   email.Address
}

func NewSender(mailer Mailer, fromEmail string) *Sender {
	return &Sender{
		mailer: mailer,
		from: email.Address{
			Email: fromEmail,
			Name:  "Confa",
		},
	}
}

func (s *Sender) Send(ctx context.Context, message *sender.Message, to ...email.Address) error {
	var msg email.Email
	var err error
	switch pl := message.Message.(type) { // nolint: gocritic
	case *sender.Message_LoginViaEmail_:
		msg, err = newLoginViaEmail(
			s.from,
			to,
			pl.LoginViaEmail.SecretLoginUrl,
		)
	case *sender.Message_TalkRecordingReady_:
		msg, err = newTalkRecordingReady(
			s.from,
			to,
			pl.TalkRecordingReady.ConfaUrl,
			pl.TalkRecordingReady.ConfaTitle,
			pl.TalkRecordingReady.TalkUrl,
			pl.TalkRecordingReady.TalkTitle,
		)
	default:
		return errors.New("unknown email message")
	}
	if err != nil {
		return fmt.Errorf("failed to render email: %w", err)
	}
	return s.mailer.Send(ctx, msg)
}

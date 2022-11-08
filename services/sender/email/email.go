package email

import (
	"context"
	"errors"
	"fmt"

	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/proto/sender"
)

type Sender struct {
	mailer *email.Sender
}

func NewSender(server, port, fromAddress, password string, secure bool) *Sender {
	return &Sender{
		mailer: email.NewSender(server, port, fromAddress, password, secure),
	}
}

func (s *Sender) Send(ctx context.Context, message *sender.Message, toAddress string) error {
	var msg email.Email
	var err error
	switch pl := message.Message.(type) { // nolint: gocritic
	case *sender.Message_LoginViaEmail_:
		msg, err = newLoginViaEmail(
			toAddress,
			pl.LoginViaEmail.SecretLoginUrl,
		)
	case *sender.Message_TalkRecordingReady_:
		msg, err = newTalkRecordingReady(
			toAddress,
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

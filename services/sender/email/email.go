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
	switch message := message.Message.(type) { // nolint: gocritic
	case *sender.Message_LoginViaEmail_:
		msg, err := newLoginViaEmail(toAddress, message.LoginViaEmail.SecretLoginUrl)
		if err != nil {
			return fmt.Errorf("failed to render email: %w", err)
		}
		return s.mailer.Send(ctx, msg)
	}

	return errors.New("unknown message")
}

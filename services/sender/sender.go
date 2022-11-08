package sender

import (
	"context"
	"errors"
	"fmt"

	"github.com/aromancev/confa/internal/proto/iam"
	"github.com/aromancev/confa/internal/proto/sender"
)

type EmailSender interface {
	Send(ctx context.Context, message *sender.Message, toAddress string) error
}

type Sender struct {
	email     EmailSender
	iamClient iam.IAM
}

func NewSender(email EmailSender, iamClient iam.IAM) *Sender {
	return &Sender{
		email:     email,
		iamClient: iamClient,
	}
}

func (s *Sender) Send(ctx context.Context, send *sender.Send) error {
	switch delivery := send.Delivery.Delivery.(type) {
	case *sender.Delivery_Auto_:
		// For now automatic delivery only supports email, but in future this should select what platform to use.
		// Possibly from user notification settings or something.
		user, err := s.iamClient.GetUser(ctx, &iam.UserLookup{
			UserId: delivery.Auto.UserId,
		})
		if err != nil {
			return fmt.Errorf("failed to find user for automatic delivery: %w", err)
		}
		address, ok := findIdent(user.Idents, iam.Platform_EMAIL)
		if !ok {
			return errors.New("user does not have email identificator")
		}
		err = s.email.Send(ctx, send.Message, address)
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
		return nil
	case *sender.Delivery_Email_:
		err := s.email.Send(ctx, send.Message, delivery.Email.ToAddress)
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
		return nil
	default:
		return errors.New("unknown delivery method")
	}
}

func findIdent(idents []*iam.User_Ident, platform iam.Platform) (string, bool) {
	for _, ident := range idents {
		if ident.Platform == platform {
			return ident.Value, true
		}
	}
	return "", false
}

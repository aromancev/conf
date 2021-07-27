package auth

import (
	"time"

	"github.com/aromancev/confa/internal/platform/email"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

const (
	emailExpire = 24 * time.Hour
	apiExpire   = 15 * time.Minute
)

type EmailClaims struct {
	jwt.StandardClaims
	Address string `json:"adr"`
}

func NewEmailClaims(address string) *EmailClaims {
	return &EmailClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(emailExpire).Unix(),
		},
		Address: address,
	}
}

func (c EmailClaims) Valid() error {
	if err := c.StandardClaims.Valid(); err != nil {
		return err
	}
	return email.Validate(c.Address)
}

type Account int

const (
	AccountGuest Account = 0
	AccountUser  Account = 1
	AccountAdmin Account = 2
)

type APIClaims struct {
	jwt.StandardClaims
	UserID  uuid.UUID `json:"uid"`
	Account Account   `json:"acc"`
}

func NewAPIClaims(userID uuid.UUID, acc Account) *APIClaims {
	return &APIClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(apiExpire).Unix(),
		},
		UserID:  userID,
		Account: acc,
	}
}

func (c APIClaims) Valid() error {
	return c.StandardClaims.Valid()
}

func (c APIClaims) ExpiresIn() time.Duration {
	return apiExpire
}

func (c APIClaims) AllowedWrite() bool {
	return c.Account == AccountUser || c.Account == AccountAdmin
}

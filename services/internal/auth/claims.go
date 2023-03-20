package auth

import (
	"time"

	"github.com/aromancev/confa/internal/platform/email"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

const (
	emailExpire    = 2 * time.Hour
	apiExpire      = 15 * time.Minute
	guestAPIExpire = 24 * time.Hour
)

type EmailClaims struct {
	jwt.StandardClaims
	ID      string `json:"id"`
	Address string `json:"adr"`
}

func NewEmailClaims(address string) EmailClaims {
	return EmailClaims{
		ID: uuid.NewString(),
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

func NewAPIClaims(userID uuid.UUID, acc Account) APIClaims {
	return APIClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(apiExpire).Unix(),
		},
		UserID:  userID,
		Account: acc,
	}
}

func NewGuesAPIClaims(userID uuid.UUID) *APIClaims {
	return &APIClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(guestAPIExpire).Unix(),
		},
		UserID:  userID,
		Account: AccountGuest,
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

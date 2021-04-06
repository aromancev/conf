package auth

import (
	"crypto/ecdsa"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"

	"github.com/aromancev/confa/internal/platform/email"
)

const (
	algorithm   = "ES256"
	emailExpire = 24 * time.Hour
	UserExpire  = 15 * time.Minute
)

type EmailClaims struct {
	jwt.StandardClaims
	Address string `json:"adr"`
}

func (c EmailClaims) Valid() error {
	if err := c.StandardClaims.Valid(); err != nil {
		return err
	}
	if err := email.ValidateEmail(c.Address); err != nil {
		return err
	}
	return nil
}

type UserClaims struct {
	jwt.StandardClaims
	UserID uuid.UUID `json:"uid"`
}

type userClaims struct {
	jwt.StandardClaims
	UserID string `json:"uid"`
}

func (c userClaims) Valid() error {
	if err := c.StandardClaims.Valid(); err != nil {
		return err
	}
	if _, err := uuid.Parse(c.UserID); err != nil {
		return err
	}
	return nil
}

type Signer struct {
	key    *ecdsa.PrivateKey
	method jwt.SigningMethod
}

func NewSigner(secretKey string) (*Signer, error) {
	key, err := jwt.ParseECPrivateKeyFromPEM([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &Signer{
		key:    key,
		method: jwt.GetSigningMethod(algorithm),
	}, nil
}

func (s *Signer) EmailToken(address string) (string, error) {
	if err := email.ValidateEmail(address); err != nil {
		return "", err
	}

	now := time.Now()
	token := jwt.NewWithClaims(s.method, EmailClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(emailExpire).Unix(),
			IssuedAt:  now.Unix(),
		},
		Address: address,
	})
	signed, err := token.SignedString(s.key)
	if err != nil {
		return "", err
	}
	return signed, nil
}

func (s *Signer) UserToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(s.method, UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(UserExpire).Unix(),
			IssuedAt:  now.Unix(),
		},
		UserID: userID,
	})
	signed, err := token.SignedString(s.key)
	if err != nil {
		return "", err
	}
	return signed, nil
}

type Verifier struct {
	key *ecdsa.PublicKey
}

func NewVerifier(publicKey string) (*Verifier, error) {
	key, err := jwt.ParseECPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		return nil, err
	}
	return &Verifier{
		key: key,
	}, nil
}

func (v *Verifier) EmailToken(token string) (EmailClaims, error) {
	var claims EmailClaims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return v.key, nil
	})
	if err != nil {
		return EmailClaims{}, err
	}
	if !parsed.Valid {
		return EmailClaims{}, errors.New("invalid token")
	}
	return claims, nil
}

func (v *Verifier) UserToken(token string) (UserClaims, error) {
	var claims userClaims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return v.key, nil
	})
	if err != nil {
		return UserClaims{}, err
	}
	if !parsed.Valid {
		return UserClaims{}, errors.New("invalid token")
	}

	userID, _ := uuid.Parse(claims.UserID)
	return UserClaims{
		StandardClaims: claims.StandardClaims,
		UserID:         userID,
	}, nil
}

func Authenticate(r *http.Request) (uuid.UUID, error) {
	id, _ := uuid.Parse("28164069-5ec3-405b-a9cc-641cf29588ed") //todo: Unhardcode this. ONLY FOR TESTING
	return id, nil
}

func Bearer(r *http.Request) (string, error) {
	rawToken := r.Header.Get("Authorization")

	authArray := strings.Split(rawToken, " ")
	if len(authArray) < 2 {
		return "", errors.New("wrong header format")
	}

	bearer, token := authArray[0], authArray[1]
	if bearer != "Bearer" {
		return "", errors.New("wrong type")
	}

	return token, nil
}

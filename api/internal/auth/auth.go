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
	algorithm = "ES256"

	emailExpire  = 24 * time.Hour
	accessExpire = 15 * time.Minute
)

type EmailClaims struct {
	jwt.StandardClaims
	Address string `json:"adr"`
}

func (c EmailClaims) Valid() error {
	if err := c.StandardClaims.Valid(); err != nil {
		return err
	}
	if err := email.Validate(c.Address); err != nil {
		return err
	}
	return nil
}

type APIClaims struct {
	jwt.StandardClaims
	UserID uuid.UUID `json:"uid"`
}

func (c APIClaims) Valid() error {
	if err := c.StandardClaims.Valid(); err != nil {
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
	now := time.Now()
	token := jwt.NewWithClaims(s.method, EmailClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(emailExpire).Unix(),
		},
		Address: address,
	})
	signed, err := token.SignedString(s.key)
	if err != nil {
		return "", err
	}
	return signed, nil
}

func (s *Signer) AccessToken(userID uuid.UUID) (string, time.Duration, error) {
	now := time.Now()
	token := jwt.NewWithClaims(s.method, APIClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(accessExpire).Unix(),
		},
		UserID: userID,
	})
	signed, err := token.SignedString(s.key)
	if err != nil {
		return "", 0, err
	}
	return signed, accessExpire, nil
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

func (v *Verifier) Verify(token string, claims jwt.Claims) error {
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return v.key, nil
	})
	if err != nil {
		return err
	}
	if !parsed.Valid {
		return errors.New("invalid token")
	}
	return nil
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

func (v *Verifier) AccessToken(token string) (APIClaims, error) {
	var claims APIClaims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return v.key, nil
	})
	if err != nil {
		return APIClaims{}, err
	}
	if !parsed.Valid {
		return APIClaims{}, errors.New("invalid token")
	}

	return claims, nil
}

func Bearer(r *http.Request) string {
	token := r.Header.Get("Authorization")
	parts := strings.Split(token, " ")
	if len(parts) < 2 {
		return ""
	}
	bearer, token := parts[0], parts[1]
	if bearer != "Bearer" {
		return ""
	}
	return token
}

func Session(r *http.Request) string {
	session, err := r.Cookie(sessionKey)
	if err != nil {
		return ""
	}
	return session.Value
}

func SetSession(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionKey,
		Value:    value,
		HttpOnly: true,
	})
}

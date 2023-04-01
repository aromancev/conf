package gsi

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Creds struct {
	ClientID     string
	ClientSecret string
}

type ID struct {
	Email      string
	GivenName  string
	FamilyName string
	Picture    string
}

type IDAuth struct {
	publicKey *rsa.PublicKey
	clientID  string
}

func NewIDAuth(pk *rsa.PublicKey, clientID string) *IDAuth {
	return &IDAuth{
		publicKey: pk,
		clientID:  clientID,
	}
}

func (a *IDAuth) Verify(token string) (ID, error) {
	var claims idClaims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(*jwt.Token) (interface{}, error) {
		return a.publicKey, nil
	})
	if err != nil {
		return ID{}, err
	}
	if !parsed.Valid {
		return ID{}, errors.New("invalid token")
	}
	if claims.Audience != a.clientID {
		return ID{}, errors.New("invalid audience")
	}
	return ID{
		Email:      claims.Email,
		GivenName:  claims.GivenName,
		FamilyName: claims.FamilyName,
		Picture:    claims.Picture,
	}, nil
}

type PublicKey struct {
	client *http.Client
}

func NewPublicKey(client *http.Client) *PublicKey {
	return &PublicKey{
		client: client,
	}
}

func (k PublicKey) PEM(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	const url = "https://www.googleapis.com/oauth2/v1/certs"

	if kid == "" {
		return nil, errors.New("invalid key id")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code")
	}
	var certs map[string]string
	err = json.NewDecoder(resp.Body).Decode(&certs)
	if err != nil {
		return nil, err
	}
	keyStr, ok := certs[kid]
	if !ok {
		return nil, errors.New("certificate for this key not found")
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(keyStr))
	if err != nil {
		return nil, err
	}
	return key, nil
}

func KeyID(token string) string {
	i := strings.Index(token, ".")
	if i == -1 {
		return ""
	}
	header := token[:i]
	buf, err := base64.RawStdEncoding.DecodeString(header)
	if err != nil {
		return ""
	}
	var payload struct {
		KeyID string `json:"kid"`
	}
	err = json.Unmarshal(buf, &payload)
	if err != nil {
		return ""
	}
	return payload.KeyID
}

type idClaims struct {
	ExpiresAt  int64  `json:"exp,omitempty"`
	Audience   string `json:"aud"`
	Issuer     string `json:"iss"`
	Email      string `json:"email"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
}

func (c idClaims) Valid() error {
	if c.Issuer != "accounts.google.com" && c.Issuer != "https://accounts.google.com" {
		return errors.New("invalid issuer")
	}
	if c.ExpiresAt < time.Now().Unix() {
		return errors.New("token expired")
	}
	return nil
}

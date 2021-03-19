package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	secretKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIB8fmVWhMdAo/UkDNN4UGo8PYwKxz/lN7nilmYa2KEkboAoGCCqGSM49
AwEHoUQDQgAETrMd0Br7GOpE7US1jJ7LbL0L8vIi3NxRxnXhOxDWaAhd4MxdF17f
AY5OGjJpPdWJ8TDMQH7Es98SAB9pVRVZhg==
-----END EC PRIVATE KEY-----`
	publicKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETrMd0Br7GOpE7US1jJ7LbL0L8vIi
3NxRxnXhOxDWaAhd4MxdF17fAY5OGjJpPdWJ8TDMQH7Es98SAB9pVRVZhg==
-----END PUBLIC KEY-----`
)

func TestEmailToken(t *testing.T) {
	t.Parallel()

	sign, err := NewSigner(secretKey)
	require.NoError(t, err)
	verify, err := NewVerifier(publicKey)
	require.NoError(t, err)

	t.Run("Happy path", func(t *testing.T) {
		email := "test@test.com"
		token, err := sign.EmailToken("test@test.com")
		require.NoError(t, err)
		claims, err := verify.EmailToken(token)
		require.NoError(t, err)
		assert.Equal(t, email, claims.Address)
		assert.NotZero(t, claims.ExpiresAt)
		assert.NotZero(t, claims.IssuedAt)
	})

	t.Run("Invalid signature returns error", func(t *testing.T) {
		token, err := sign.EmailToken("test@test.com")
		require.NoError(t, err)
		parts := strings.Split(token, ".")

		// Replacing the email in the token
		body, _ := base64.StdEncoding.DecodeString(parts[1])
		var claims EmailClaims
		_ = json.Unmarshal(body, &claims)
		claims.Address = "test@test.com"
		body, _ = json.Marshal(claims)
		parts[2] = base64.StdEncoding.EncodeToString(body)
		token = strings.Join(parts, ".")

		_, err = verify.EmailToken(token)
		assert.Error(t, err)
	})
}

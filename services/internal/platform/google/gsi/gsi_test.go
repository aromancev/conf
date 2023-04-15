package gsi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyID(t *testing.T) {
	assert.Equal(
		t,
		"986ee9a3b7520b494df54fe32e3e5c4ca685c89d",
		KeyID("eyJhbGciOiJSUzI1NiIsImtpZCI6Ijk4NmVlOWEzYjc1MjBiNDk0ZGY1NGZlMzJlM2U1YzRjYTY4NWM4OWQiLCJ0eXAiOiJKV1QifQ.stub.stub"),
	)
}

func TestPublicKey(t *testing.T) {
	const resp = `{
		"test": "-----BEGIN CERTIFICATE-----\nMIIDJjCCAg6gAwIBAgIIJReUdLXYTr0wDQYJKoZIhvcNAQEFBQAwNjE0MDIGA1UE\nAwwrZmVkZXJhdGVkLXNpZ25vbi5zeXN0ZW0uZ3NlcnZpY2VhY2NvdW50LmNvbTAe\nFw0yMzAzMzExNTIyNDJaFw0yMzA0MTcwMzM3NDJaMDYxNDAyBgNVBAMMK2ZlZGVy\nYXRlZC1zaWdub24uc3lzdGVtLmdzZXJ2aWNlYWNjb3VudC5jb20wggEiMA0GCSqG\nSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCvni13eFO/zsjBQ2F1z5tgsifLi0FVxqy9\n1J3uVskguDnwLgMnREk9zNd3+uV/PNga+Cm3c6R/9qcl3lpqHX3o/durBUN16VwN\ngCG5qMHOfjRCM4E95+90PnNKjXyLs60bueECFFIQZ7o+POlyTfACiCphwOCQXUHN\nOxEH4OTGmuxiGnmmYvlGdf7oRg/m3ab2Mn7+g/2/XK9mRPlQ9vYjA6TkYOWVE9u+\nn5olb9EzXwhNTeohuTBJOzWAkYVY7uPCfFPRAFoUPxrxp6/W2ZHnR6Er5LPY5G2+\n5YHFvNOpdcvf2qA0lpjBnJb7bTjS++5mdoauJtzFPze3dwmVB0zHAgMBAAGjODA2\nMAwGA1UdEwEB/wQCMAAwDgYDVR0PAQH/BAQDAgeAMBYGA1UdJQEB/wQMMAoGCCsG\nAQUFBwMCMA0GCSqGSIb3DQEBBQUAA4IBAQAht4VFQNv03bibZqmmUvw+um999dJ5\nKm1rtTQKLhXVnZUceU53W1NsGmw00SzGhC1CoUeWeLCKnmxfexB6soeqB2hagRzb\nQAUz6fJM/YIhnfeqgrkk5VzcWA4Lr0V2Q4lYuprwqS5A/iFOmOgeg8Kx4fPOs72v\n2zp8Ekg9stihLkkG/+eL2lPjzIke7VsmuWvB8MRDO2jyJXmU8xn8A2smHbFNaRf3\np1dK2ndlqMUwKTuIBjiJLKXTUbFaZDvNng70PYZdu3s0L3eUgEmfkpKCJl74f6yi\nNChi9RIe882AFzqGhAhXwfnvRkntDnakGpZR6/S0RVavbH7V97/DEm9B\n-----END CERTIFICATE-----\n"
	}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	pk, err := NewPublicKey(server.URL, server.Client())
	require.NoError(t, err)
	_, err = pk.PEM(context.Background(), "test")
	assert.NoError(t, err)
}

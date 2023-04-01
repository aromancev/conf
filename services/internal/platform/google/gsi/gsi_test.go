package gsi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
		"1aae8d7c92058b5eea456895bf89084571e306f3": "-----BEGIN CERTIFICATE-----
		MIIDJjCCAg6gAwIBAgIIJhH74ArYaJMwDQYJKoZIhvcNAQEFBQAwNjE0MDIGA1UE
		AwwrZmVkZXJhdGVkLXNpZ25vbi5zeXN0ZW0uZ3NlcnZpY2VhY2NvdW50LmNvbTAe
		Fw0yMzAzMjMxNTIyNDFaFw0yMzA0MDkwMzM3NDFaMDYxNDAyBgNVBAMMK2ZlZGVy
		YXRlZC1zaWdub24uc3lzdGVtLmdzZXJ2aWNlYWNjb3VudC5jb20wggEiMA0GCSqG
		SIb3DQEBAQUAA4IBDwAwggEKAoIBAQCyC4goi+URUGwQaTvuJXbI1DGlj9CSLLfK
		4x9jjCmac+V6+UMoBK7q4/8Ia5vNOGIEeUKFgMM29h86K1ZcPCnFsn8yqZqMD50N
		usjkt2CDImlKhY8pOE8nUIpGFGJFcmNcaLoaDpN9thHC7TOLIOBlnXbk2zO405Q7
		WlzWoa6fhI+J/Njs5joG0ANkOpNYcVN+b/KGAGDISUT6vh35mo9721hioKcC8gsG
		5ls9i3dQLd8Cv0mW11Q7ni/EpuWjbPZ934f3MGf2NFk4GS5VUHxBJXmyt74fbcJM
		EAFKlml7RKsmleAb0C5XgNrsVEcEL0D6gtgt75QXn/lzr9x3tz8pAgMBAAGjODA2
		MAwGA1UdEwEB/wQCMAAwDgYDVR0PAQH/BAQDAgeAMBYGA1UdJQEB/wQMMAoGCCsG
		AQUFBwMCMA0GCSqGSIb3DQEBBQUAA4IBAQApnzWvLCszOvoKTfnP1v4p6dhQ8ZBl
		Vlep9TqxftEyFez0iMPjuWr1xxuDAKB65AwaeN92cZ0S47MAMew1C9mTcmBpbOfN
		rJhgnnV28DpAyOdNA1lBlNXT6CkH755cHh13mccfW1oD7BQI1iebPZtVnw1DUpyk
		UwjPU8jZ/S4co0Ykyagr0IJzO/hRDk7wRJFyy13wBRaeIXKxM/lUdPASRDlt9Blz
		BYdO3kt8m+RqDPbpWX/WB0M3xinVnfigFZpbfeiD4FF+evdaFl88w7kxg0NJLrFL
		5NK55IfnxpKGhxW011KrZn8YR2hpTsbuasq4/p4nrvI/qOzmN/nul+bA
		-----END CERTIFICATE-----"
	}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(resp))
	}))

	pk := NewPublicKey(server.Client())
	_, err := pk.PEM(context.Background(), "1aae8d7c92058b5eea456895bf89084571e306f3")
	assert.NoError(t, err)
}

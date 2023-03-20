package web

import (
	"net/http"
	"time"

	"github.com/aromancev/confa/internal/auth"
)

const (
	guestAPIExpire = 24 * time.Hour
)

type Auth struct {
	domain string
}

func NewAuth(domain string) *Auth {
	return &Auth{
		domain: domain,
	}
}

func (a *Auth) Token(request *http.Request) string {
	return auth.NewHTTPContext(request).Token()
}

func (a *Auth) Session(request *http.Request) string {
	session, err := request.Cookie(sessionCookie)
	if err != nil {
		return ""
	}
	return session.Value
}

func (a *Auth) GuestToken(request *http.Request) string {
	token, err := request.Cookie(guestCookie)
	if err != nil {
		return ""
	}
	return token.Value
}

func (a *Auth) SetSession(writer http.ResponseWriter, sessionKey string) {
	http.SetCookie(writer, &http.Cookie{
		Name:     sessionCookie,
		Value:    sessionKey,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Domain:   a.domain,
	})
}

func (a *Auth) ResetSession(writer http.ResponseWriter) {
	http.SetCookie(writer, &http.Cookie{
		Name:    sessionCookie,
		Value:   "",
		Expires: time.Unix(0, 0),
	})
}

func (a *Auth) SetGuestToken(writer http.ResponseWriter, token string) {
	http.SetCookie(writer, &http.Cookie{
		Name:     guestCookie,
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Domain:   a.domain,
		Expires:  time.Now().Add(guestAPIExpire),
		MaxAge:   int(guestAPIExpire.Seconds()),
	})
}

func (a *Auth) ResetGuestToken(writer http.ResponseWriter) {
	http.SetCookie(writer, &http.Cookie{
		Name:    guestCookie,
		Value:   "",
		Expires: time.Unix(0, 0),
	})
}

const (
	sessionCookie = "session"
	guestCookie   = "guest-token"
)

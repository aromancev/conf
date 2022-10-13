package auth

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type Context interface {
	Token() string
	SetSession(value string)
	Session() string
	ResetSession()
	SetGuestClaims(value string)
	GuestClaims() string
	ResetGuestClaims()
}

type HTTPContext struct {
	writer  http.ResponseWriter
	request *http.Request
}

func NewHTTPContext(w http.ResponseWriter, r *http.Request) *HTTPContext {
	return &HTTPContext{
		writer:  w,
		request: r,
	}
}

func (c *HTTPContext) SetSession(value string) {
	http.SetCookie(c.writer, &http.Cookie{
		Name:     sessionKey,
		Value:    value,
		HttpOnly: true,
	})
}

func (c *HTTPContext) ResetSession() {
	http.SetCookie(c.writer, &http.Cookie{
		Name:     sessionKey,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
}

func (c *HTTPContext) Session() string {
	session, err := c.request.Cookie(sessionKey)
	if err != nil {
		return ""
	}
	return session.Value
}

func (c *HTTPContext) SetGuestClaims(value string) {
	http.SetCookie(c.writer, &http.Cookie{
		Name:     guestClaimsKey,
		Value:    value,
		HttpOnly: true,
		Expires:  time.Now().Add(guestAPIExpire),
		MaxAge:   int(guestAPIExpire.Seconds()),
	})
}

func (c *HTTPContext) ResetGuestClaims() {
	http.SetCookie(c.writer, &http.Cookie{
		Name:     guestClaimsKey,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
}

func (c *HTTPContext) GuestClaims() string {
	claims, err := c.request.Cookie(guestClaimsKey)
	if err != nil {
		return ""
	}
	return claims.Value
}

func (c *HTTPContext) Token() string {
	token := c.request.Header.Get("Authorization")
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

type WSockContext struct {
	writer  http.ResponseWriter
	request *http.Request
}

func NewWSockContext(w http.ResponseWriter, r *http.Request) *WSockContext {
	return &WSockContext{
		writer:  w,
		request: r,
	}
}

func (c *WSockContext) SetSession(value string) {
	http.SetCookie(c.writer, &http.Cookie{
		Name:     sessionKey,
		Value:    value,
		HttpOnly: true,
	})
}

func (c *WSockContext) ResetSession() {
	http.SetCookie(c.writer, &http.Cookie{
		Name:     sessionKey,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
}

func (c *WSockContext) Session() string {
	session, err := c.request.Cookie(sessionKey)
	if err != nil {
		return ""
	}
	return session.Value
}

func (c *WSockContext) SetGuestClaims(value string) {
	http.SetCookie(c.writer, &http.Cookie{
		Name:     guestClaimsKey,
		Value:    value,
		HttpOnly: true,
		Expires:  time.Now().Add(guestAPIExpire),
		MaxAge:   int(guestAPIExpire.Seconds()),
	})
}

func (c *WSockContext) ResetGuestClaims() {
	http.SetCookie(c.writer, &http.Cookie{
		Name:     guestClaimsKey,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
}

func (c *WSockContext) GuestClaims() string {
	claims, err := c.request.Cookie(guestClaimsKey)
	if err != nil {
		return ""
	}
	return claims.Value
}

func (c *WSockContext) Token() string {
	t, ok := c.request.URL.Query()["t"]
	if !ok {
		return ""
	}
	if len(t) != 1 {
		return ""
	}
	return t[0]
}

func SetContext(parent context.Context, ctx Context) context.Context {
	return context.WithValue(parent, ctxKey{}, ctx)
}

func Ctx(ctx context.Context) Context {
	val := ctx.Value(ctxKey{})
	if c, ok := val.(Context); ok {
		return c
	}
	panic("auth.Context not set")
}

const (
	sessionKey     = "session"
	guestClaimsKey = "guest-claims"
)

type ctxKey struct{}

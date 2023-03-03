package auth

import (
	"context"
	"net/http"
	"strings"
)

type Context interface {
	Token() string
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

type HTTPContext struct {
	request *http.Request
}

func NewHTTPContext(r *http.Request) *HTTPContext {
	return &HTTPContext{
		request: r,
	}
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
	request *http.Request
}

func NewWSockContext(r *http.Request) *WSockContext {
	return &WSockContext{
		request: r,
	}
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

type ctxKey struct{}

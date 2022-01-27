package main

import (
	"net/http"

	"github.com/movio/bramble"
)

func init() {
	bramble.RegisterPlugin(&PassHeaders{})
}

type PassHeaders struct {
	bramble.BasePlugin
}

func (p *PassHeaders) ID() string {
	return "pass-headers"
}

func (p *PassHeaders) ApplyMiddlewarePublicMux(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		for key, values := range r.Header {
			for _, value := range values {
				ctx = bramble.AddOutgoingRequestsHeaderToContext(ctx, key, value)
			}
		}
		h.ServeHTTP(rw, r.WithContext(ctx))
	})
}

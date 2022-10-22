package main

import (
	"net/http"

	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/movio/bramble"
)

type PassHeaders struct {
	bramble.BasePlugin
}

func (*PassHeaders) ID() string {
	return "pass-headers"
}

func (*PassHeaders) ApplyMiddlewarePublicMux(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		for key, values := range r.Header {
			for _, value := range values {
				ctx = bramble.AddOutgoingRequestsHeaderToContext(ctx, key, value)
			}
		}
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

type TraceHeader struct {
	bramble.BasePlugin
}

func (*TraceHeader) ID() string {
	return "trace-header"
}

func (*TraceHeader) ApplyMiddlewarePublicMux(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, traceID := trace.Ctx(r.Context())
		ctx = bramble.AddOutgoingRequestsHeaderToContext(ctx, traceHeader, traceID)
		w.Header().Set(traceHeader, traceID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

const (
	traceHeader = "Trace-Id"
)

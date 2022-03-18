package web

import (
	"net/http"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/internal/platform/trace"
)

func withHTTPAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := auth.SetContext(r.Context(), auth.NewHTTPContext(w, r))
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withWebSocketAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := auth.SetContext(r.Context(), auth.NewWSockContext(w, r))
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withNewTrace(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.Ctx(r.Context())
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

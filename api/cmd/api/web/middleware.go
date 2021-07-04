package web

import (
	"net/http"

	"github.com/aromancev/confa/internal/auth"
)

func withAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := auth.SetContext(r.Context(), auth.NewHTTPContext(w, r))
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	router http.Handler
}

func NewHandler(media http.Handler) *Handler {
	r := httprouter.New()

	r.GET("/health", ok)

	r.GET(
		"/v1/:media_id/:file",
		serveMedia(media),
	)
	return &Handler{
		router: r,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, traceID := trace.Ctx(r.Context())
	w.Header().Set("Trace-Id", traceID)

	defer func() {
		if err := recover(); err != nil {
			log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Msg("ServeHTTP panic")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	lw := newResponseWriter(w)
	r = r.WithContext(ctx)
	h.router.ServeHTTP(lw, r)

	lw.Event(ctx, r).Msg("HTTP served")
}

type responseWriter struct {
	http.ResponseWriter
	code int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, code: http.StatusOK}
}

func (w *responseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Event(ctx context.Context, r *http.Request) *zerolog.Event {
	var event *zerolog.Event
	if w.code >= http.StatusInternalServerError {
		event = log.Ctx(ctx).Error()
	} else {
		event = log.Ctx(ctx).Info()
	}
	return event.Str("method", r.Method).Int("code", w.code).Str("url", r.URL.String())
}

func ok(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	_, _ = w.Write([]byte("OK"))
}

func serveMedia(media http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		r.URL.Path = ps.ByName("media_id") + "/" + ps.ByName("file")
		media.ServeHTTP(w, r)
	}
}

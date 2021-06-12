package plog

import (
	"bufio"
	"context"
	"net"
	"net/http"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ResponseWriter struct {
	http.ResponseWriter
	code int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{ResponseWriter: w, code: http.StatusOK}
}

func (w *ResponseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	panic("ResponseWriter does not implement http.Hijacker")
}

func (w *ResponseWriter) Event(ctx context.Context, r *http.Request) *zerolog.Event {
	var event *zerolog.Event
	if w.code >= http.StatusInternalServerError {
		event = log.Ctx(ctx).Error()
	} else {
		event = log.Ctx(ctx).Info()
	}
	return event.Str("method", r.Method).Int("code", w.code).Str("url", r.URL.String())
}

func JobEvent(ctx context.Context, job beanstalk.Job) *zerolog.Event {
	return log.Ctx(ctx).Info().Uint64("jobId", job.ID).Int("releases", job.Stats.Releases)
}

func WithContext(ctx context.Context, z zerolog.Context) context.Context {
	l := z.Logger()
	return l.WithContext(ctx)
}

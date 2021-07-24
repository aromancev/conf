package web

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/aromancev/confa/internal/platform/trace"
)

//go:generate gqlgen

type Code string

const (
	CodeInternal       = "INTERNAL"
	CodeBadRequest     = "BAD_REQUEST"
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeDuplicateEntry = "DUPLICATE_ENTRY"
	CodeNotFound       = "NOT_FOUND"
)

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Handler struct {
	router http.Handler
}

func NewHandler(resolver *Resolver) *Handler {
	r := http.NewServeMux()

	r.HandleFunc("/health", ok)
	r.Handle("/query",
		withAuth(
			handler.NewDefaultServer(
				NewExecutableSchema(Config{Resolvers: resolver}),
			),
		),
	)
	r.HandleFunc(
		"/rtc/v1/ws",
		serveRTC(resolver.upgrader, resolver.sfuAddr),
	)
	r.HandleFunc("/dev/", playground.Handler("API playground", "/api/query"))

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

	lw.Event(ctx, r).Msg("Web served")
}

func ok(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func newError(code Code, message string) *gqlerror.Error {
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
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

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	panic("ResponseWriter does not implement http.Hijacker")
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

const (
	batchLimit = 100
)

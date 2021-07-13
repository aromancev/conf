package web

import (
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rs/zerolog/log"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/aromancev/confa/internal/platform/api"
	"github.com/aromancev/confa/internal/platform/plog"
	"github.com/aromancev/confa/internal/platform/trace"
)

//go:generate gqlgen

type Code string

const (
	CodeInternal     = "INTERNAL"
	CodeInvalidParam = "INVALID_PARAM"
	CodeDuplicatedEntity = "DUPLICATED_ENTITY"
	CodeUnauthorized = "UNAUTHORIZED"
)

type Web struct {
	router http.Handler
}

func New(resolver *Resolver) *Web {
	r := http.NewServeMux()

	r.HandleFunc("/health", ok)
	r.Handle("/query",
		withAuth(
			handler.NewDefaultServer(
				NewExecutableSchema(Config{Resolvers: resolver}),
			),
		),
	)
	r.HandleFunc("/dev/", playground.Handler("API playground", "/api/query"))

	return &Web{
		router: r,
	}
}

func (h *Web) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, traceID := trace.Ctx(r.Context())
	w.Header().Set("Trace-Id", traceID)

	defer func() {
		if err := recover(); err != nil {
			log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Msg("ServeHTTP panic")
			_ = api.InternalError(w)
		}
	}()
	lw := plog.NewResponseWriter(w)
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

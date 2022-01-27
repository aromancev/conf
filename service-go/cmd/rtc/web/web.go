package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/aromancev/confa/internal/platform/trace"
)

type Code string

const (
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeDuplicateEntry   = "DUPLICATE_ENTRY"
	CodeNotFound         = "NOT_FOUND"
	CodePermissionDenied = "PERMISSION_DENIED"
	CodeUnknown          = "UNKNOWN_CODE"
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
	r.Handle(
		"/query",
		withHTTPAuth(
			&relay.Handler{
				Schema: graphql.MustParseSchema(schema, resolver, graphql.UseFieldResolvers()),
			},
		),
	)
	r.HandleFunc(
		"/room/",
		withWSockAuthFunc(
			serveRTC(resolver.rooms, resolver.publicKey, resolver.upgrader, resolver.sfuConn, resolver.producer, resolver.eventWatcher),
		),
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
	h.router.ServeHTTP(w, r.WithContext(ctx))
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

func newInternalError() *gqlerror.Error {
	return &gqlerror.Error{
		Message: "internal system error",
		Extensions: map[string]interface{}{
			"code": CodeUnknown,
		},
	}
}

const (
	batchLimit = 100
)

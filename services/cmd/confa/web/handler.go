package web

import (
	"fmt"
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/platform/trace"
)

type Handler struct {
	router http.Handler
}

func NewHandler(resolver *Resolver, publicKey *auth.PublicKey) *Handler {
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

	return &Handler{
		router: r,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := trace.New(r.Context(), r.Header.Get("Trace-Id"))

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

package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/platform/trace"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
)

type Email interface {
	AddRoutes(r *httprouter.Router)
	Emails() []email.Email
}

type Handler struct {
	router *httprouter.Router
	email  Email
}

func NewHandler(emails Email) *Handler {
	h := &Handler{
		router: httprouter.New(),
		email:  emails,
	}

	h.router.GET("/health", ok)
	h.router.GET("/state/email", h.emails)

	emails.AddRoutes(h.router)

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	traceID := r.Header.Get("Trace-Id")
	if traceID != "" {
		ctx = trace.New(ctx, traceID)
	}

	defer func() {
		if err := recover(); err != nil {
			log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Bytes("stack", debug.Stack()).Msg("ServeHTTP panic")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	h.router.ServeHTTP(w, r.WithContext(ctx))
}

func (h Handler) emails(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	_ = json.NewEncoder(w).Encode(h.email.Emails())
}

func ok(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	_, _ = w.Write([]byte("OK"))
}

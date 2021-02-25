package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/platform/trace"
)

type Handler struct {
	router    http.Handler
	confaCRUD *confa.CRUD
}

func New(confaCRUD *confa.CRUD) *Handler {
	r := httprouter.New()
	h := &Handler{
		confaCRUD: confaCRUD,
	}

	r.GET("/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = w.Write([]byte("OK"))
	})

	r.POST("confa/v1/confas", trace.Wrap(h.createConfa))

	h.router = r
	return h
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

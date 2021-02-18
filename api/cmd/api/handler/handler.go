package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	router http.Handler
}

func New() *Handler {
	router := httprouter.New()

	return &Handler{router: router}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

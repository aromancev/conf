package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func ok(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	_, _ = w.Write([]byte("OK"))
}

func serveMedia(media http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		r.URL.Path = ps.ByName("media_id") + "/" + ps.ByName("file")
		media.ServeHTTP(w, r)
	}
}

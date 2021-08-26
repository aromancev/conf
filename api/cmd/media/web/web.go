package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/aromancev/confa/proto/queue"
	"github.com/google/uuid"

	"github.com/prep/beanstalk"

	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	mediaDir string
	router   http.Handler
	producer *beanstalk.Producer
}

func NewHandler(media http.Handler, mediaDir string, producer *beanstalk.Producer) *Handler {
	r := httprouter.New()

	r.GET("/health", ok)
	r.GET("/v1/:media_id/:file", serveMedia(media))
	r.POST("/v1/upload/image", uploadImage(mediaDir, producer))

	return &Handler{
		mediaDir: mediaDir,
		router:   r,
		producer: producer,
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

func uploadImage(mediaDir string, producer *beanstalk.Producer) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// FormFile returns the first file for the given key `myFile`
		// it also returns the FileHeader so we can get the Filename,
		// the Header and the size of the file

		file, _, err := r.FormFile("img")
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("No img header")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		mediaID := uuid.New().String()
		err = os.Mkdir(path.Join(mediaDir, mediaID), os.ModePerm)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg(fmt.Sprintf("Error creating directory: %s", path.Join(mediaDir, mediaID)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fo, err := os.Create(path.Join(mediaDir, mediaID, "img.raw"))
		if err != nil {
			log.Ctx(ctx).Err(err).Msg(fmt.Sprintf("Error creating file inside directory: %s", path.Join(mediaDir, mediaID)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer fo.Close()

		_, err = io.Copy(fo, file)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg(fmt.Sprintf("File read error: %s", mediaID))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		payload, err := queue.Marshal(&queue.ImageJob{MediaId: mediaID}, trace.ID(ctx))
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Error marshalling payload")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = producer.Put(ctx, queue.TubeImage, payload, beanstalk.PutParams{TTR: 15})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Queue error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

package handler

import (
	"context"
	"fmt"
	"github.com/aromancev/confa/internal/confa/talk"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/platform/api"
	"github.com/aromancev/confa/internal/platform/backoff"
	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/platform/plog"
	"github.com/aromancev/confa/internal/platform/trace"
)

const (
	TubeEmail = "email"
)

const (
	jobRetries = 10
)

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Handler struct {
	baseURL   string
	router    http.Handler
	confaCRUD *confa.CRUD
	talkCRUD  *talk.CRUD
	sender    *email.Sender
	producer  Producer
	sign      *auth.Signer
	verify    *auth.Verifier
}

func New(baseURL string, confaCRUD *confa.CRUD, talkCRUD *talk.CRUD, sender *email.Sender, producer Producer, sign *auth.Signer, verify *auth.Verifier) *Handler {
	r := httprouter.New()
	h := &Handler{
		baseURL:   baseURL,
		confaCRUD: confaCRUD,
		talkCRUD:  talkCRUD,
		sender:    sender,
		producer:  producer,
		sign:      sign,
		verify:    verify,
	}

	r.GET("/iam/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = w.Write([]byte("OK"))
	})
	r.GET("/confa/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = w.Write([]byte("OK"))
	})

	r.POST("/confa/v1/confas", h.createConfa)
	r.GET("/confa/v1/confas/:confa_id", h.confa)

	r.POST("/confa/v1/confas/:confa_id/talks", h.createTalk)
	r.GET("/confa/v1/talks/:talk_id", h.talk)

	r.POST("/iam/v1/login", h.login)

	h.router = r
	return h
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, _ := trace.Ctx(r.Context())

	defer func() {
		if err := recover(); err != nil {
			log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Msg("ServeHTTP panic")
			_ = api.InternalError().Write(ctx, w)
		}
	}()
	lw := plog.NewResponseWriter(w)
	r = r.WithContext(ctx)
	h.router.ServeHTTP(lw, r)

	lw.Event(ctx, r).Msg("HTTP served")
}

func (h *Handler) ServeJob(ctx context.Context, job *beanstalk.Job) {
	ctx, _ = trace.Job(ctx, job)

	plog.JobEvent(ctx, *job).Msg("Job received")

	defer func() {
		if err := recover(); err != nil {
			log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Msg("ServeJob panic")
		}
	}()

	var err error
	switch job.Stats.Tube {
	case TubeEmail:
		err = h.sendEmail(ctx, job)
	default:
		err = fmt.Errorf("unknown tube: %s", job.Stats.Tube)
	}
	if err != nil {
		if job.Stats.Releases >= jobRetries {
			log.Ctx(ctx).Err(err).Msg("Job retries exceeded. Burying.")
			if err := job.Bury(ctx); err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to bury job")
			}
			return
		}

		bo := backoff.Backoff{
			Factor: 1.2,
			Min:    10 * time.Second,
			Max:    10 * time.Minute,
		}
		log.Ctx(ctx).Err(err).Msg("Job failed. Releasing.")
		if err := job.ReleaseWithParams(ctx, job.Stats.Priority, bo.ForAttempt(float64(job.Stats.Releases))); err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to release job")
		}
		return
	}

	if err := job.Delete(ctx); err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to delete job")
	}

	plog.JobEvent(ctx, *job).Msg("Job served")
}

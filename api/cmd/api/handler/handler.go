package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/aromancev/confa/internal/user/session"

	"github.com/aromancev/confa/internal/user/ident"

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
	baseURL     string
	router      http.Handler
	confaCRUD   *confa.CRUD
	talkCRUD    *talk.CRUD
	sessionCRUD *session.CRUD
	identCRUD   *ident.CRUD
	sender      *email.Sender
	producer    Producer
	sign        *auth.Signer
	verify      *auth.Verifier
}

func New(baseURL string, confaCRUD *confa.CRUD, talkCRUD *talk.CRUD, sessionCRUD *session.CRUD, identCRUD *ident.CRUD, sender *email.Sender, producer Producer, signer *auth.Signer, verifier *auth.Verifier) *Handler {
	r := httprouter.New()
	h := &Handler{
		baseURL:     baseURL,
		confaCRUD:   confaCRUD,
		talkCRUD:    talkCRUD,
		sessionCRUD: sessionCRUD,
		identCRUD:   identCRUD,
		sender:      sender,
		producer:    producer,
		sign:        signer,
		verify:      verifier,
	}

	withTrace := func(h httprouter.Handle) httprouter.Handle {
		return trace.WriteHeader(h)
	}

	r.GET("/iam/health", ok)
	r.POST(
		"/iam/v1/login",
		withTrace(login(baseURL, signer, producer)),
	)
	r.POST(
		"/iam/v1/sessions",
		withTrace(createSession(verifier, signer, identCRUD, sessionCRUD)),
	)
	r.GET(
		"/iam/v1/token",
		withTrace(createToken(signer, sessionCRUD)),
	)

	r.GET("/confa/health", ok)
	r.POST(
		"/confa/v1/confas",
		withTrace(createConfa(verifier, confaCRUD)),
	)
	r.GET(
		"/confa/v1/confas/:confa_id",
		withTrace(getConfa(confaCRUD)),
	)
	r.POST(
		"/confa/v1/confas/:confa_id/talks",
		withTrace(createTalk(verifier, talkCRUD)),
	)
	r.GET(
		"/confa/v1/talks/:talk_id",
		withTrace(getTalk(talkCRUD)),
	)

	h.router = r
	return h
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, _ := trace.Ctx(r.Context())

	defer func() {
		if err := recover(); err != nil {
			log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Msg("ServeHTTP panic")
			_ = api.InternalError().Write(w)
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

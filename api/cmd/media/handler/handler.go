package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/internal/media/video"
	"github.com/aromancev/confa/internal/platform/api"
	"github.com/aromancev/confa/internal/platform/backoff"
	"github.com/aromancev/confa/internal/platform/plog"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/proto/queue"
)

const (
	jobRetries = 1
)

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type JobHandle func(ctx context.Context, job *beanstalk.Job) error

type HTTP struct {
	router http.Handler
}

func NewHTTP(media http.Handler) *HTTP {
	r := httprouter.New()

	r.GET("/health", ok)

	r.GET(
		"/v1/:media_id/:file",
		serveMedia(media),
	)
	return &HTTP{
		router: r,
	}
}

func (h *HTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	lw.Event(ctx, r).Msg("HTTP served")
}

type Job struct {
	route func(job *beanstalk.Job) JobHandle
}

func NewJob(converter *video.Converter) *Job {
	return &Job{
		route: func(job *beanstalk.Job) JobHandle {
			switch job.Stats.Tube {
			case queue.TubeVideo:
				return processVideo(converter)
			default:
				return nil
			}
		},
	}
}

func (h *Job) ServeJob(ctx context.Context, job *beanstalk.Job) {
	j, err := queue.Unmarshal(job.Body)
	if err != nil {
		log.Ctx(ctx).Error().Str("tube", job.Stats.Tube).Msg("No handle for job. Burying.")
		return
	}
	ctx = trace.New(ctx, j.TraceId)
	job.Body = j.Payload

	plog.JobEvent(ctx, *job).Msg("Job received")

	defer func() {
		if err := recover(); err != nil {
			log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Msg("ServeJob panic")
		}
	}()

	handle := h.route(job)
	if handle == nil {
		log.Ctx(ctx).Error().Str("tube", job.Stats.Tube).Msg("No handle for job. Burying.")
		return
	}

	err = handle(ctx, job)
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

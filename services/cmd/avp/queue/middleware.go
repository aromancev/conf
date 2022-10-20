package queue

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
)

// AutoTouch performs periodic job touch to prevent it from being scheduled to another worker.
// IMPORTANT: Make sure the handler never hangs. Otherwise this middleware will touch the job infinetely.
func AutoTouch(h JobHandle) JobHandle {
	return func(ctx context.Context, job *beanstalk.Job) {
		touchCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		var touch sync.WaitGroup
		touch.Add(1)
		go func() {
			defer touch.Done()

			for {
				select {
				case <-time.After(job.TouchAfter()):
					err := job.Touch(touchCtx)
					switch {
					case errors.Is(err, context.Canceled):
						return
					case errors.Is(err, beanstalk.ErrJobFinished):
						return
					case err != nil:
						log.Ctx(ctx).Err(err).Msg("Failed to touch job.")
						return
					}
					log.Ctx(ctx).Debug().Uint64("jobId", job.ID).Msg("Touched job.")
				case <-touchCtx.Done():
					return
				}
			}
		}()

		h(ctx, job)
		cancel()
		touch.Wait()
		log.Ctx(ctx).Debug().Msg("Auto touch finished.")
	}
}

package tracker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var (
	ErrClosed = errors.New("runtime stopped")
)

type Tracker interface {
	Close() error
}

type NewTracker func(ctx context.Context, roomID uuid.UUID) (Tracker, error)

type Runtime struct {
	mutex          sync.Mutex
	trackersClosed sync.WaitGroup
	trackers       map[trackerKey]*trackerEntry
	shuttingDown   bool
	ctx            context.Context
}

func NewRuntime() *Runtime {
	return &Runtime{
		trackers: map[trackerKey]*trackerEntry{},
		ctx:      context.Background(),
	}
}

// StartTracker returns true if a tracker with the same room and role already exists and live.
// If tracker creation errored before and was not clean up by GC yet, it will try to initiate it again.
// Context passed to `newTracker` will be cancelled only when tracker is closed. It is safe to use for background processing.
func (r *Runtime) StartTracker(ctx context.Context, roomID uuid.UUID, role string, expireAt time.Time, newTracker NewTracker) (bool, error) {
	key := trackerKey{
		roomID: roomID,
		role:   role,
	}
	r.mutex.Lock()
	if r.shuttingDown {
		r.mutex.Unlock()
		return false, ErrClosed
	}
	entry, ok := r.trackers[key]
	if ok {
		entry.expireAt = expireAt
	} else {
		trackerCtx, cancel := context.WithCancel(r.ctx)
		entry = &trackerEntry{
			status:    statusInit,
			expireAt:  expireAt,
			ctx:       trackerCtx,
			cancelCtx: cancel,
		}
	}
	r.trackers[key] = entry
	r.trackersClosed.Add(1)
	r.mutex.Unlock()

	entry.Lock()
	defer entry.Unlock()
	if atomic.LoadUint32(&entry.status) == statusLive {
		return true, nil
	}
	tracker, err := newTracker(entry.ctx, roomID)
	if err != nil {
		atomic.StoreUint32(&entry.status, statusError)
		return false, err
	}
	entry.tracker = tracker
	atomic.StoreUint32(&entry.status, statusLive)
	return false, nil
}

// Run starts the GC cycle to remove and close expired peer.
// NOT SAFE TO CALL CONCURRENTLY.
func (r *Runtime) Run(ctx context.Context, gcPeriod time.Duration) {
	r.mutex.Lock()
	r.ctx = ctx
	r.mutex.Unlock()
	for {
		select {
		case <-ctx.Done():
			r.mutex.Lock()
			r.shuttingDown = true
			r.closeTrackers(ctx, r.trackers)
			r.mutex.Unlock()
			r.trackersClosed.Wait()
			return
		case <-time.After(gcPeriod):
		}

		r.mutex.Lock()
		now := time.Now()
		total := len(r.trackers)
		toClose := map[trackerKey]*trackerEntry{}
		for key, entry := range r.trackers {
			if atomic.LoadUint32(&entry.status) != statusError && entry.expireAt.After(now) {
				continue
			}
			toClose[key] = entry
			delete(r.trackers, key)
		}
		r.mutex.Unlock()

		r.closeTrackers(ctx, toClose)
		log.Ctx(ctx).Debug().Int("trackersTotal", total).Int("trackersClosed", len(toClose)).Msg("Tracker runtime GC cycle.")
	}
}

const (
	statusInit uint32 = iota
	statusLive
	statusClosing
	statusClosed
	statusError
)

type trackerEntry struct {
	sync.Mutex
	ctx       context.Context
	cancelCtx func()
	// Only atomic access be ause it's used in GC cycle without lock toa void blocking the whole map.
	status  uint32
	tracker Tracker
	// Can only be changed and read when `Runtime.mutex` is locked.
	expireAt time.Time
}

type trackerKey struct {
	roomID uuid.UUID
	role   string
}

func (r *Runtime) closeTrackers(ctx context.Context, trackers map[trackerKey]*trackerEntry) {
	if len(trackers) == 0 {
		return
	}
	for _, entry := range trackers {
		entry := entry
		go func() {
			entry.cancelCtx()
			switch atomic.LoadUint32(&entry.status) {
			case statusError, statusClosing, statusClosed:
				return
			}

			entry.Lock()
			defer entry.Unlock()
			atomic.StoreUint32(&entry.status, statusClosing)
			if err := entry.tracker.Close(); err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to close tracker.")
			}
			atomic.StoreUint32(&entry.status, statusClosed)
			r.trackersClosed.Done()
		}()
	}
}

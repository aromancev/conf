package event

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type SharedWatcher struct {
	watcher      Watcher
	m            sync.RWMutex
	maxEvents    int64
	rooms        map[uuid.UUID]*sharedRoom
	shuttingDown uint32
	serveCtx     context.Context
}

func NewSharedWatcher(w Watcher, maxEvents int64) *SharedWatcher {
	sw := &SharedWatcher{
		watcher:   w,
		maxEvents: maxEvents,
		rooms:     make(map[uuid.UUID]*sharedRoom),
	}
	// Locking right after creation because Serve must be called before use.
	sw.m.Lock()
	return sw
}

func (w *SharedWatcher) Watch(_ context.Context, roomID uuid.UUID) (Cursor, error) {
	if atomic.LoadUint32(&w.shuttingDown) != 0 {
		return nil, ErrShuttingDown
	}

	room, err := w.getOrCreate(roomID).GetOrStart(roomID, w.maxEvents)
	if err != nil {
		return nil, err
	}
	return room.SharedCursor(), nil
}

// Serve sets the context for all new room watches and starts the GC cycle.
// GC cycle will clean up unused rooms.
func (w *SharedWatcher) Serve(ctx context.Context, gcPeriod time.Duration) error {
	w.serveCtx = ctx
	w.m.Unlock()

	var candidates []*sharedRoom
	for {
		if atomic.LoadUint32(&w.shuttingDown) != 0 {
			return ErrShuttingDown
		}

		candidates = candidates[:0]
		w.m.RLock()
		log.Ctx(ctx).Debug().Int("len", len(w.rooms)).Msg("Shared watcher total rooms.")
		for _, room := range w.rooms {
			if room.OpenedCursors() == 0 {
				candidates = append(candidates, room)
			}
		}
		w.m.RUnlock()

		select {
		case <-time.After(gcPeriod):
		case <-ctx.Done():
			w.Shutdown(ctx)
			return ctx.Err()
		}

		if len(candidates) == 0 {
			log.Ctx(ctx).Debug().Msg("Shared watcher GC cycle completed (no rooms removed).")
			continue
		}
		w.m.Lock()
		for _, room := range candidates {
			if room.OpenedCursors() != 0 {
				continue
			}
			delete(w.rooms, room.ID())
			go room.Close(ctx)
			log.Ctx(ctx).Debug().Str("roomId", room.ID().String()).Msg("Room removed from shared watcher.")
		}
		w.m.Unlock()
		log.Ctx(ctx).Debug().Msg("Shared watcher GC cycle completed.")
	}
}

func (w *SharedWatcher) Shutdown(ctx context.Context) {
	atomic.CompareAndSwapUint32(&w.shuttingDown, 0, 1)

	w.m.Lock()
	defer w.m.Unlock()

	var wg sync.WaitGroup
	wg.Add(len(w.rooms))
	for _, r := range w.rooms {
		go func(room *sharedRoom) {
			err := room.Close(ctx)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to close room.")
			}
			wg.Done()
		}(r)
	}
	w.rooms = make(map[uuid.UUID]*sharedRoom)
	wg.Wait()
}

func (w *SharedWatcher) Len() int {
	w.m.RLock()
	defer w.m.RUnlock()
	return len(w.rooms)
}

func (w *SharedWatcher) getOrCreate(roomID uuid.UUID) *sharedRoom {
	// Try with read lock to reduce contention.
	w.m.RLock()
	if r, ok := w.rooms[roomID]; ok {
		w.m.RUnlock()
		return r
	}
	w.m.RUnlock()

	// Not found, will have to acquire write lock.
	w.m.Lock()
	defer w.m.Unlock()
	// Check again in case it was created between lock calls.
	if r, ok := w.rooms[roomID]; ok {
		return r
	}
	r := newSharedRoom(w.serveCtx, w.watcher)
	w.rooms[roomID] = r
	return r
}

type sharedRoom struct {
	m        sync.Mutex
	iterator sync.WaitGroup
	watcher  Watcher
	watching uint32
	room     *Room
	roomID   uuid.UUID
	ctx      context.Context
	cancel   func()
}

func newSharedRoom(ctx context.Context, w Watcher) *sharedRoom {
	ctx, cancel := context.WithCancel(ctx)
	return &sharedRoom{
		watcher: w,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (r *sharedRoom) GetOrStart(roomID uuid.UUID, maxEvents int64) (*Room, error) {
	r.m.Lock()
	defer r.m.Unlock()

	if r.room != nil {
		return r.room, nil
	}

	r.iterator.Add(1)
	cur, err := r.watcher.Watch(r.ctx, roomID)
	if err != nil {
		return nil, err
	}

	room := NewRoom(cur, maxEvents)
	go func() {
		log.Ctx(r.ctx).Debug().Str("roomId", roomID.String()).Msg("Room iteration started.")
		for {
			_, err := room.Next(r.ctx)
			if errors.Is(err, ErrCursorClosed) || errors.Is(err, context.Canceled) {
				break
			}
			if err != nil {
				log.Ctx(r.ctx).Err(err).Str("roomId", roomID.String()).Msg("Room iteration failed.")
				return
			}
		}
		r.iterator.Done()
		log.Ctx(r.ctx).Debug().Str("roomId", roomID.String()).Msg("Room iteration finished.")
	}()

	r.room = room
	r.roomID = roomID
	atomic.CompareAndSwapUint32(&r.watching, 0, 1)
	return room, nil
}

func (r *sharedRoom) OpenedCursors() int64 {
	if atomic.LoadUint32(&r.watching) == 0 {
		return 0
	}
	return r.room.OpenedCursors()
}

func (r *sharedRoom) ID() uuid.UUID {
	return r.roomID
}

func (r *sharedRoom) Close(ctx context.Context) error {
	r.cancel()
	r.iterator.Wait()
	return r.room.Close(ctx)
}

type Room struct {
	m                       sync.RWMutex
	cursor                  Cursor
	head, tail              *eventNode
	openedCursors, len, cap int64
	closeOnce               sync.Once
}

func NewRoom(cursor Cursor, maxEvents int64) *Room {
	seed := newEventNode(Event{})
	return &Room{
		cursor: cursor,
		head:   seed,
		tail:   seed,
		cap:    maxEvents,
	}
}

func (r *Room) Next(ctx context.Context) (Event, error) {
	ev, err := r.cursor.Next(ctx)

	// Not locking before because Next can hang and prevent all cursors from iterating.
	r.m.Lock()
	defer r.m.Unlock()

	if err != nil {
		_ = r.close(ctx)
		return Event{}, err
	}

	r.head = r.head.SetNext(ev)
	l := atomic.LoadInt64(&r.len)
	if l+1 <= r.cap {
		atomic.AddInt64(&r.len, 1)
	} else {
		r.tail.Evict()
		r.tail = r.tail.next
	}

	return ev, nil
}

func (r *Room) SharedCursor() *SharedCursor {
	r.m.RLock()
	defer r.m.RUnlock()
	return newSharedCursor(r.head, &r.openedCursors)
}

func (r *Room) OpenedCursors() int64 {
	return atomic.LoadInt64(&r.openedCursors)
}

func (r *Room) Len() int64 {
	return atomic.LoadInt64(&r.len)
}

func (r *Room) Close(ctx context.Context) error {
	r.m.Lock()
	defer r.m.Unlock()
	return r.close(ctx)
}

func (r *Room) close(ctx context.Context) error {
	r.closeOnce.Do(func() {
		for node := r.tail; node != nil; node = node.next {
			node.Evict()
		}
		r.head.ResetNext()
	})
	return r.cursor.Close(ctx)
}

type SharedCursor struct {
	node          *eventNode
	closeOnce     sync.Once
	openedCursors *int64
	closed        chan struct{}
}

func newSharedCursor(node *eventNode, opened *int64) *SharedCursor {
	atomic.AddInt64(opened, 1)
	return &SharedCursor{
		node:          node,
		openedCursors: opened,
		closed:        make(chan struct{}),
	}
}

func (c *SharedCursor) Next(ctx context.Context) (Event, error) {
	if c.node.IsEvicted() {
		_ = c.Close(ctx)
		return Event{}, ErrCursorClosed
	}

	select {
	case <-c.node.HasNext():
		// Check evicted again in case it was evicted while waiting on HasNext.
		if c.node.IsEvicted() {
			_ = c.Close(ctx)
			return Event{}, ErrCursorClosed
		}
		next, err := c.node.Next()
		if err != nil {
			_ = c.Close(ctx)
			return Event{}, err
		}
		c.node = next
		return next.event, nil
	case <-c.closed:
		_ = c.Close(ctx)
		return Event{}, ErrCursorClosed
	case <-ctx.Done():
		_ = c.Close(ctx)
		return Event{}, ctx.Err()
	}
}

func (c *SharedCursor) Close(_ context.Context) error {
	c.closeOnce.Do(func() {
		close(c.closed)
		atomic.AddInt64(c.openedCursors, -1)
	})
	return nil
}

type eventNode struct {
	hasNext chan struct{}
	event   Event
	next    *eventNode
	evicted uint32
}

func newEventNode(ev Event) *eventNode {
	node := &eventNode{
		hasNext: make(chan struct{}),
		event:   ev,
	}
	return node
}

func (e *eventNode) HasNext() <-chan struct{} {
	return e.hasNext
}

func (e *eventNode) SetNext(ev Event) *eventNode {
	node := newEventNode(ev)
	e.next = node
	close(e.hasNext)
	return node
}

// Next is not safe to call before HasNext returned.
func (e *eventNode) Next() (*eventNode, error) {
	if e.next == nil {
		return nil, ErrCursorClosed
	}
	return e.next, nil
}

func (e *eventNode) ResetNext() {
	close(e.hasNext)
}

func (e *eventNode) Evict() {
	atomic.CompareAndSwapUint32(&e.evicted, 0, 1)
}

func (e *eventNode) IsEvicted() bool {
	return atomic.LoadUint32(&e.evicted) != 0
}

package double

import (
	"context"
	"sync"

	"github.com/aromancev/confa/internal/event"
	"github.com/google/uuid"
)

type FakeCursor struct {
	events    <-chan event.Event
	closed    chan struct{}
	closeOnce sync.Once
}

func (c *FakeCursor) Next(ctx context.Context) (event.Event, error) {
	select {
	case ev, ok := <-c.events:
		if !ok {
			return event.Event{}, event.ErrCursorClosed
		}
		return ev, nil
	case <-ctx.Done():
		return event.Event{}, ctx.Err()
	case <-c.closed:
		panic("reading from closed cursor")
	}
}

func (c *FakeCursor) Close(ctx context.Context) error {
	c.closeOnce.Do(func() {
		close(c.closed)
	})
	return nil
}

type FakeWatcher struct {
	events chan event.Event
}

func NewFakeWatcher(events ...event.Event) *FakeWatcher {
	ch := make(chan event.Event, len(events))
	for _, e := range events {
		ch <- e
	}
	return &FakeWatcher{
		events: ch,
	}
}

func (w *FakeWatcher) Watch(ctx context.Context, roomID uuid.UUID) (event.Cursor, error) {
	return &FakeCursor{
		events: w.events,
		closed: make(chan struct{}),
	}, nil
}

func (w *FakeWatcher) Put(e event.Event) {
	w.events <- e
}

func (w *FakeWatcher) Close() {
	close(w.events)
}

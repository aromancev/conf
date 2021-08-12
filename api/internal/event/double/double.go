package double

import (
	"context"

	"github.com/aromancev/confa/internal/event"
	"github.com/google/uuid"
)

type FakeCursor struct {
	events <-chan event.Event
}

func (c *FakeCursor) Next(ctx context.Context) (event.Event, error) {
	ev, ok := <-c.events
	if !ok {
		return event.Event{}, event.ErrCursorClosed
	}
	return ev, nil
}

func (c *FakeCursor) Close(ctx context.Context) error {
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
	return &FakeCursor{events: w.events}, nil
}

func (w *FakeWatcher) Put(e event.Event) {
	w.events <- e
}

func (w *FakeWatcher) Close() {
	close(w.events)
}

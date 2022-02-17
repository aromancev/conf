package double

import (
	"context"
	"errors"
	"testing"

	"github.com/aromancev/confa/event"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFakeWatcher(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	events := []event.Event{
		{
			ID: uuid.New(),
		},
		{
			ID: uuid.New(),
		},
		{
			ID: uuid.New(),
		},
	}

	watcher := NewFakeWatcher(events...)
	watcher.Close()
	cur, err := watcher.Watch(ctx, uuid.New())
	require.NoError(t, err)
	fetched := make([]event.Event, 0, len(events))
	for {
		ev, err := cur.Next(ctx)
		if errors.Is(err, event.ErrCursorClosed) {
			break
		}
		fetched = append(fetched, ev)
	}
	assert.Equal(t, events, fetched)
}

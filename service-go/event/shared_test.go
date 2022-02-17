package event_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/event/double"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoom(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		t.Parallel()

		events := []event.Event{
			{ID: uuid.New()},
			{ID: uuid.New()},
			{ID: uuid.New()},
		}
		watcher := double.NewFakeWatcher(events...)
		defer watcher.Close()

		c, err := watcher.Watch(ctx, uuid.New())
		require.NoError(t, err)
		room := event.NewRoom(c, 3)
		defer room.Close(ctx)

		go func() {
			for i := 0; i < len(events); i++ {
				_, _ = room.Next(ctx)
			}
		}()

		fetched := make([]event.Event, 0, len(events))
		cur := room.SharedCursor()
		for i := 0; i < len(events); i++ {
			ev, err := cur.Next(ctx)
			require.NoError(t, err)
			fetched = append(fetched, ev)
		}
		assert.Equal(t, events, fetched)
	})
	t.Run("Blocks when drained", func(t *testing.T) {
		t.Parallel()

		created := event.Event{ID: uuid.New()}

		watcher := double.NewFakeWatcher(created)
		cur, err := watcher.Watch(ctx, uuid.New())
		require.NoError(t, err)
		room := event.NewRoom(cur, 1)

		var done sync.WaitGroup
		done.Add(1)
		go func() {
			// Trying to iterate async (since this will hang).
			cur := room.SharedCursor()
			fetched, err := cur.Next(ctx)
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
			done.Done()
		}()

		// Waiting for a bit to ensure the iterator is blocked and not just quit.
		time.Sleep(time.Second)
		// Inserting the event, which should unblock the cursor.
		_, _ = room.Next(ctx)
		// Waiting for cursor to finish the check.
		done.Wait()
	})
	t.Run("Unblocks when context cancelled", func(t *testing.T) {
		t.Parallel()

		watcher := double.NewFakeWatcher()
		cur, err := watcher.Watch(ctx, uuid.New())
		require.NoError(t, err)
		room := event.NewRoom(cur, 1)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		_, err = room.SharedCursor().Next(ctx)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
	t.Run("Unblocks when cursor closed", func(t *testing.T) {
		t.Parallel()

		watcher := double.NewFakeWatcher()
		wc, err := watcher.Watch(ctx, uuid.New())
		require.NoError(t, err)
		room := event.NewRoom(wc, 1)

		cur := room.SharedCursor()
		go func() {
			time.Sleep(time.Second) // Giving the cursor time to lock.
			cur.Close(ctx)
		}()
		_, err = cur.Next(ctx)
		assert.ErrorIs(t, err, event.ErrCursorClosed)
		assert.Zero(t, room.OpenedCursors())
	})
	t.Run("Cursor closes once", func(t *testing.T) {
		t.Parallel()

		watcher := double.NewFakeWatcher()
		c, err := watcher.Watch(ctx, uuid.New())
		require.NoError(t, err)
		room := event.NewRoom(c, 1)

		cur := room.SharedCursor()
		assert.Equal(t, int64(1), room.OpenedCursors())
		require.NoError(t, cur.Close(ctx))
		assert.Zero(t, room.OpenedCursors())
		require.NoError(t, cur.Close(ctx))
		assert.Zero(t, room.OpenedCursors())
	})
	t.Run("Closed room closes all cursors", func(t *testing.T) {
		t.Parallel()

		watcher := double.NewFakeWatcher()
		c, err := watcher.Watch(ctx, uuid.New())
		require.NoError(t, err)
		room := event.NewRoom(c, 1)

		cur := room.SharedCursor()
		_ = room.Close(ctx)
		_, err = cur.Next(ctx)
		require.ErrorIs(t, err, event.ErrCursorClosed)
		assert.Zero(t, room.OpenedCursors())
	})
	t.Run("Room does not grow beyond capacity", func(t *testing.T) {
		t.Parallel()

		watcher := double.NewFakeWatcher(event.Event{}, event.Event{})
		c, err := watcher.Watch(ctx, uuid.New())
		require.NoError(t, err)
		room := event.NewRoom(c, 1)
		assert.Equal(t, int64(0), room.Len())

		_, err = room.Next(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), room.Len())
		_, err = room.Next(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), room.Len())
	})
	t.Run("Slow cursors get evicted", func(t *testing.T) {
		t.Parallel()

		watcher := double.NewFakeWatcher(event.Event{}, event.Event{})
		c, err := watcher.Watch(ctx, uuid.New())
		require.NoError(t, err)
		room := event.NewRoom(c, 1)

		shouldExpire := room.SharedCursor()
		_, err = room.Next(ctx)
		require.NoError(t, err)
		shouldBeOk := room.SharedCursor()
		_, err = room.Next(ctx)
		require.NoError(t, err)

		_, err = shouldExpire.Next(ctx)
		assert.ErrorIs(t, err, event.ErrCursorClosed)
		_, err = shouldBeOk.Next(ctx)
		assert.NoError(t, err)
	})
	t.Run("Concurrent cursors fetch correct events", func(t *testing.T) {
		t.Parallel()

		const (
			numEvents  = 100
			numCursors = 1000
		)

		watcher := double.NewFakeWatcher()
		defer watcher.Close()

		cursor, err := watcher.Watch(ctx, uuid.New())
		require.NoError(t, err)

		room := event.NewRoom(cursor, numEvents)
		defer room.Close(ctx)

		// Starting concurrent cursors.
		results := make(chan []event.Event, numCursors)
		var cursorsDone, cursorsCreated sync.WaitGroup
		cursorsDone.Add(numCursors)
		cursorsCreated.Add(numCursors)
		for i := 0; i < numCursors; i++ {
			go func() {
				cur := room.SharedCursor()
				cursorsCreated.Done()

				fetched := make([]event.Event, 0, numCursors)
				for j := 0; j < numEvents; j++ {
					ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
					ev, err := cur.Next(ctx)
					switch {
					case errors.Is(err, event.ErrCursorClosed):
						panic(fmt.Sprintf("cursor closed too soon on iteration: %d", j))
					case err != nil:
						panic("unexpected error: " + err.Error())
					}
					fetched = append(fetched, ev)
					cancel()
				}
				results <- fetched
				_ = cur.Close(ctx)
				cursorsDone.Done()
			}()
		}

		// All cursors should be created before iterating the room.
		// Otherwise some cursors will miss an event.
		cursorsCreated.Wait()

		// Iterating over room so it pulls events from the shared cursor.
		go func() {
			for i := 0; i < numEvents; i++ {
				_, err = room.Next(ctx)
			}
		}()

		// Populating random events.
		expected := make([]event.Event, numEvents)
		for i := 0; i < numEvents; i++ {
			expected[i] = event.Event{ID: uuid.New()}
		}

		// Inserting events into the shared cursor.
		go func() {
			for i := 0; i < numEvents; i++ {
				watcher.Put(expected[i])
			}
		}()

		// Waiting for everything to finish.
		cursorsDone.Wait()
		close(results)

		// Checking that all concurrent cursors received all events in correct order.
		for fetched := range results {
			assert.Equal(t, expected, fetched)
		}
		assert.Zero(t, room.OpenedCursors())
	})
}

func TestSharedWatcher(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		t.Parallel()

		created := event.Event{ID: uuid.New()}
		watcher := double.NewFakeWatcher(created)
		defer watcher.Close()

		shared := event.NewSharedWatcher(watcher, 10)
		go shared.Serve(ctx, time.Second) // nolint: errcheck
		cur, err := shared.Watch(ctx, uuid.New())
		require.NoError(t, err)

		fetched, err := cur.Next(ctx)
		require.NoError(t, err)
		assert.Equal(t, created, fetched)
	})

	t.Run("GC cleans up unused rooms", func(t *testing.T) {
		t.Parallel()

		watcher := double.NewFakeWatcher()
		defer watcher.Close()

		shared := event.NewSharedWatcher(watcher, 10)
		go shared.Serve(ctx, 10*time.Millisecond) // nolint: errcheck

		cur, err := shared.Watch(ctx, uuid.New())
		require.NoError(t, err)

		time.Sleep(time.Second)
		require.Equal(t, 1, shared.Len()) // One cursor opened.

		cur.Close(ctx)

		time.Sleep(time.Second)
		require.Equal(t, 0, shared.Len()) // No cursors opened.
	})

	t.Run("Shutdown closes cursors", func(t *testing.T) {
		t.Parallel()

		watcher := double.NewFakeWatcher()
		defer watcher.Close()

		shared := event.NewSharedWatcher(watcher, 10)
		go shared.Serve(ctx, time.Second) // nolint: errcheck
		cur, err := shared.Watch(ctx, uuid.New())
		require.NoError(t, err)

		go func() {
			// Give the cursor some time to lock.
			time.Sleep(time.Second)
			shared.Shutdown(ctx)
		}()

		_, err = cur.Next(ctx)
		require.ErrorIs(t, err, event.ErrCursorClosed)
	})
}

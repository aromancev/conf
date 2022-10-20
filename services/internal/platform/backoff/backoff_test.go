package backoff

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBackoff(t *testing.T) {
	t.Parallel()

	assertBetween := func(t *testing.T, actual, low, high time.Duration) {
		t.Helper()
		if actual < low {
			t.Fatalf("Got %s, Expecting >= %s", actual, low)
		}
		if actual > high {
			t.Fatalf("Got %s, Expecting <= %s", actual, high)
		}
	}

	t.Run("case 1", func(t *testing.T) {
		b := &Backoff{
			Min:    100 * time.Millisecond,
			Max:    10 * time.Second,
			Factor: 2,
		}

		assert.Equal(t, b.Duration(), 100*time.Millisecond)
		assert.Equal(t, b.Duration(), 200*time.Millisecond)
		assert.Equal(t, b.Duration(), 400*time.Millisecond)
		b.Reset()
		assert.Equal(t, b.Duration(), 100*time.Millisecond)
	})
	t.Run("case 2", func(t *testing.T) {
		b := &Backoff{
			Min:    100 * time.Millisecond,
			Max:    10 * time.Second,
			Factor: 1.5,
		}

		assert.Equal(t, b.Duration(), 100*time.Millisecond)
		assert.Equal(t, b.Duration(), 150*time.Millisecond)
		assert.Equal(t, b.Duration(), 225*time.Millisecond)
		b.Reset()
		assert.Equal(t, b.Duration(), 100*time.Millisecond)
	})
	t.Run("case 3", func(t *testing.T) {
		b := &Backoff{
			Min:    100 * time.Nanosecond,
			Max:    10 * time.Second,
			Factor: 1.75,
		}

		assert.Equal(t, b.Duration(), 100*time.Nanosecond)
		assert.Equal(t, b.Duration(), 175*time.Nanosecond)
		assert.Equal(t, b.Duration(), 306*time.Nanosecond)
		b.Reset()
		assert.Equal(t, b.Duration(), 100*time.Nanosecond)
	})
	t.Run("case 4", func(t *testing.T) {
		b := &Backoff{
			Min:    500 * time.Second,
			Max:    100 * time.Second,
			Factor: 1,
		}

		assert.Equal(t, b.Duration(), b.Max)
	})
	t.Run("for attempt", func(t *testing.T) {
		b := &Backoff{
			Min:    100 * time.Millisecond,
			Max:    10 * time.Second,
			Factor: 2,
		}

		assert.Equal(t, b.ForAttempt(0), 100*time.Millisecond)
		assert.Equal(t, b.ForAttempt(1), 200*time.Millisecond)
		assert.Equal(t, b.ForAttempt(2), 400*time.Millisecond)
		b.Reset()
		assert.Equal(t, b.ForAttempt(0), 100*time.Millisecond)
	})
	t.Run("get attempt", func(t *testing.T) {
		b := &Backoff{
			Min:    100 * time.Millisecond,
			Max:    10 * time.Second,
			Factor: 2,
		}
		assert.Equal(t, b.Attempt(), float64(0))
		assert.Equal(t, b.Duration(), 100*time.Millisecond)
		assert.Equal(t, b.Attempt(), float64(1))
		assert.Equal(t, b.Duration(), 200*time.Millisecond)
		assert.Equal(t, b.Attempt(), float64(2))
		assert.Equal(t, b.Duration(), 400*time.Millisecond)
		assert.Equal(t, b.Attempt(), float64(3))
		b.Reset()
		assert.Equal(t, b.Attempt(), float64(0))
		assert.Equal(t, b.Duration(), 100*time.Millisecond)
		assert.Equal(t, b.Attempt(), float64(1))
	})
	t.Run("jitter", func(t *testing.T) {
		b := &Backoff{
			Min:    100 * time.Millisecond,
			Max:    10 * time.Second,
			Factor: 2,
			Jitter: true,
		}

		assert.Equal(t, b.Duration(), 100*time.Millisecond)
		assertBetween(t, b.Duration(), 100*time.Millisecond, 200*time.Millisecond)
		assertBetween(t, b.Duration(), 100*time.Millisecond, 400*time.Millisecond)
		b.Reset()
		assert.Equal(t, b.Duration(), 100*time.Millisecond)
	})
	t.Run("copy", func(t *testing.T) {
		b := &Backoff{
			Min:    100 * time.Millisecond,
			Max:    10 * time.Second,
			Factor: 2,
		}
		b2 := b.Copy()
		assert.Equal(t, b, b2)
	})
	t.Run("concurrent", func(t *testing.T) {
		b := &Backoff{
			Min:    100 * time.Millisecond,
			Max:    10 * time.Second,
			Factor: 2,
		}

		wg := &sync.WaitGroup{}

		test := func() {
			time.Sleep(b.Duration())
			wg.Done()
		}

		wg.Add(2)
		go test()
		go test()
		wg.Wait()
	})
}

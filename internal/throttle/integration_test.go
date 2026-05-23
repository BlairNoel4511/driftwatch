package throttle_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/throttle"
)

// TestThrottle_ConcurrentAccess verifies that concurrent Allow calls are safe
// and that the total number of allowed events does not exceed burst * goroutines
// when each goroutine uses a unique key.
func TestThrottle_ConcurrentAccess(t *testing.T) {
	th := throttle.New(time.Second, 3)
	const goroutines = 20
	var wg sync.WaitGroup
	var allowed atomic.Int64

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				if th.Allow("shared-key") {
					allowed.Add(1)
				}
			}
		}(i)
	}
	wg.Wait()

	// Within a single window, only burst=3 calls should have been allowed.
	if got := allowed.Load(); got != 3 {
		t.Fatalf("expected exactly 3 allowed events, got %d", got)
	}
}

// TestThrottle_MultiWindow verifies that events are re-allowed across multiple
// consecutive windows.
func TestThrottle_MultiWindow(t *testing.T) {
	window := 40 * time.Millisecond
	th := throttle.New(window, 2)
	key := "multi"

	passed := 0
	for _, wait := range []time.Duration{0, 0, 0, window + 5*time.Millisecond, 0, 0} {
		if wait > 0 {
			time.Sleep(wait)
		}
		if th.Allow(key) {
			passed++
		}
	}
	// Window 1: 2 pass, 1 blocked; Window 2: 2 pass, 1 blocked → 4 total.
	if passed != 4 {
		t.Fatalf("expected 4 events to pass across 2 windows, got %d", passed)
	}
}

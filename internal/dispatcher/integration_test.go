package dispatcher_test

import (
	"sync/atomic"
	"testing"

	"github.com/yourorg/driftwatch/internal/dispatcher"
	"github.com/yourorg/driftwatch/internal/watcher"
)

// TestDispatcher_FallbackNotCalledWhenRouteMatches verifies that the fallback
// is suppressed when at least one route fires.
func TestDispatcher_FallbackNotCalledWhenRouteMatches(t *testing.T) {
	var fallbackCount int32
	d := dispatcher.New(func(e watcher.Event) { atomic.AddInt32(&fallbackCount, 1) })
	d.Register("/etc", func(e watcher.Event) {})

	for i := 0; i < 10; i++ {
		d.Dispatch(makeEvent("/etc/hosts"))
	}

	if atomic.LoadInt32(&fallbackCount) != 0 {
		t.Fatalf("fallback should not have fired, got %d calls", fallbackCount)
	}
}

// TestDispatcher_DeregisterAndRe verifies that a prefix can be removed and
// re-registered without affecting other routes.
func TestDispatcher_DeregisterAndRe(t *testing.T) {
	d := dispatcher.New(nil)
	var etcCount, varCount int32

	d.Register("/etc", func(e watcher.Event) { atomic.AddInt32(&etcCount, 1) })
	d.Register("/var", func(e watcher.Event) { atomic.AddInt32(&varCount, 1) })

	d.Dispatch(makeEvent("/etc/hosts"))
	d.Dispatch(makeEvent("/var/log/app.log"))

	d.Deregister("/etc")
	d.Dispatch(makeEvent("/etc/hosts")) // should not increment etcCount

	if atomic.LoadInt32(&etcCount) != 1 {
		t.Fatalf("expected etcCount=1, got %d", etcCount)
	}
	if atomic.LoadInt32(&varCount) != 1 {
		t.Fatalf("expected varCount=1, got %d", varCount)
	}
}

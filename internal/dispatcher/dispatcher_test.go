package dispatcher_test

import (
	"sync"
	"testing"

	"github.com/yourorg/driftwatch/internal/dispatcher"
	"github.com/yourorg/driftwatch/internal/watcher"
)

func makeEvent(path string) watcher.Event {
	return watcher.Event{Path: path}
}

func TestDispatch_MatchingPrefix(t *testing.T) {
	d := dispatcher.New(nil)
	var got []string
	d.Register("/etc", func(e watcher.Event) { got = append(got, e.Path) })

	d.Dispatch(makeEvent("/etc/nginx/nginx.conf"))

	if len(got) != 1 || got[0] != "/etc/nginx/nginx.conf" {
		t.Fatalf("expected one match, got %v", got)
	}
}

func TestDispatch_NoMatchCallsFallback(t *testing.T) {
	var fallbackCalled bool
	d := dispatcher.New(func(e watcher.Event) { fallbackCalled = true })
	d.Register("/etc", func(e watcher.Event) {})

	d.Dispatch(makeEvent("/var/log/app.log"))

	if !fallbackCalled {
		t.Fatal("expected fallback to be called")
	}
}

func TestDispatch_NoMatchNoFallback(t *testing.T) {
	d := dispatcher.New(nil)
	d.Register("/etc", func(e watcher.Event) { t.Fatal("should not be called") })
	// must not panic
	d.Dispatch(makeEvent("/tmp/file"))
}

func TestDispatch_MultipleRoutesMatch(t *testing.T) {
	d := dispatcher.New(nil)
	count := 0
	d.Register("/etc", func(e watcher.Event) { count++ })
	d.Register("/etc/nginx", func(e watcher.Event) { count++ })

	d.Dispatch(makeEvent("/etc/nginx/nginx.conf"))

	if count != 2 {
		t.Fatalf("expected 2 handlers called, got %d", count)
	}
}

func TestDeregister_RemovesRoute(t *testing.T) {
	d := dispatcher.New(nil)
	var called bool
	d.Register("/etc", func(e watcher.Event) { called = true })
	d.Deregister("/etc")

	d.Dispatch(makeEvent("/etc/hosts"))

	if called {
		t.Fatal("handler should not be called after deregister")
	}
}

func TestRegister_PanicsOnEmptyPrefix(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty prefix")
		}
	}()
	d := dispatcher.New(nil)
	d.Register("", func(e watcher.Event) {})
}

func TestRegister_PanicsOnNilHandler(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil handler")
		}
	}()
	d := dispatcher.New(nil)
	d.Register("/etc", nil)
}

func TestDispatch_ConcurrentSafe(t *testing.T) {
	d := dispatcher.New(nil)
	d.Register("/etc", func(e watcher.Event) {})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.Dispatch(makeEvent("/etc/hosts"))
		}()
	}
	wg.Wait()
}

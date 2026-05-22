package notifier_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/history"
	"github.com/yourorg/driftwatch/internal/notifier"
	"github.com/yourorg/driftwatch/internal/watcher"
)

// recordSink captures events for assertions.
type recordSink struct {
	mu     sync.Mutex
	events []watcher.DriftEvent
}

func (r *recordSink) Notify(e watcher.DriftEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
	return nil
}

func (r *recordSink) Events() []watcher.DriftEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := make([]watcher.DriftEvent, len(r.events))
	for i, e := range r.events {
		copy[i] = e
	}
	return copy
}

func TestNotifier_DispatchesToSink(t *testing.T) {
	h := history.New(10)
	sink := &recordSink{}
	n := notifier.New(h, nil, sink)

	ch := make(chan watcher.DriftEvent, 2)
	ch <- watcher.DriftEvent{Path: "/etc/hosts", Reason: "size changed"}
	ch <- watcher.DriftEvent{Path: "/etc/passwd", Reason: "mode changed"}
	close(ch)

	ctx := context.Background()
	n.Run(ctx, ch)

	events := sink.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Path != "/etc/hosts" {
		t.Errorf("unexpected path: %s", events[0].Path)
	}
}

func TestNotifier_RecordsInHistory(t *testing.T) {
	h := history.New(10)
	n := notifier.New(h, nil)

	ch := make(chan watcher.DriftEvent, 1)
	ch <- watcher.DriftEvent{Path: "/tmp/test", Reason: "missing"}
	close(ch)

	n.Run(context.Background(), ch)

	if h.Len() != 1 {
		t.Errorf("expected 1 history entry, got %d", h.Len())
	}
}

func TestNotifier_StopsOnContextCancel(t *testing.T) {
	n := notifier.New(nil, nil)
	ch := make(chan watcher.DriftEvent)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		n.Run(ctx, ch)
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run did not stop after context cancel")
	}
}

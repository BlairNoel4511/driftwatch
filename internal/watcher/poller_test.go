package watcher_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/watcher"
)

func TestPoller_EmitsDriftEvent(t *testing.T) {
	dir := t.TempDir()
	p := writeTempFile(t, dir, "app.yaml", "version: 1\n")

	w := watcher.New()
	if err := w.Snapshot([]string{p}); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	// Modify file before poller starts checking.
	if err := os.WriteFile(p, []byte("version: 2\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	poller := watcher.NewPoller(w, 20*time.Millisecond, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		poller.Run(ctx)
		close(done)
	}()

	select {
	case ev := <-w.Events:
		if ev.Path != p {
			t.Errorf("wrong path: got %s", ev.Path)
		}
	case <-ctx.Done():
		t.Fatal("timed out: no drift event received")
	}

	cancel()
	<-done
}

func TestPoller_StopsOnContextCancel(t *testing.T) {
	dir := t.TempDir()
	p := writeTempFile(t, dir, "stable.yaml", "stable: true\n")

	w := watcher.New()
	if err := w.Snapshot([]string{p}); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	poller := watcher.NewPoller(w, 50*time.Millisecond, nil)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		poller.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("poller did not stop after context cancellation")
	}
}

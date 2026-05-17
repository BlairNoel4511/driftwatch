package watcher_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/watcher"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return p
}

func TestSnapshot_And_NoDrift(t *testing.T) {
	dir := t.TempDir()
	p := writeTempFile(t, dir, "cfg.yaml", "key: value\n")

	w := watcher.New()
	if err := w.Snapshot([]string{p}); err != nil {
		t.Fatalf("Snapshot error: %v", err)
	}

	w.Check()

	select {
	case ev := <-w.Events:
		t.Errorf("unexpected drift event: %+v", ev)
	case err := <-w.Errors:
		t.Errorf("unexpected error: %v", err)
	case <-time.After(50 * time.Millisecond):
		// expected: no events
	}
}

func TestCheck_DetectsDrift(t *testing.T) {
	dir := t.TempDir()
	p := writeTempFile(t, dir, "cfg.yaml", "key: original\n")

	w := watcher.New()
	if err := w.Snapshot([]string{p}); err != nil {
		t.Fatalf("Snapshot error: %v", err)
	}

	// Mutate the file after snapshotting.
	if err := os.WriteFile(p, []byte("key: changed\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	w.Check()

	select {
	case ev := <-w.Events:
		if ev.Path != p {
			t.Errorf("expected path %s, got %s", p, ev.Path)
		}
		if ev.Expected == ev.Actual {
			t.Error("expected checksum mismatch but got equal checksums")
		}
	case err := <-w.Errors:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for drift event")
	}
}

func TestSnapshot_MissingFile(t *testing.T) {
	w := watcher.New()
	err := w.Snapshot([]string{"/nonexistent/path/cfg.yaml"})
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

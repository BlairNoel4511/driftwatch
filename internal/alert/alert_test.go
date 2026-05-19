package alert_test

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/user/driftwatch/internal/alert"
	"github.com/user/driftwatch/internal/watcher"
)

func makeSnapshot() watcher.FileState {
	return watcher.FileState{
		Size:    512,
		Mode:    0o644,
		ModTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestNotify_SizeChanged(t *testing.T) {
	var buf bytes.Buffer
	n := alert.New(&buf)

	current := makeSnapshot()
	current.Size = 1024

	ev := watcher.DriftEvent{
		Path:     "/etc/app/config.yaml",
		Snapshot: makeSnapshot(),
		Current:  &current,
	}

	if err := n.Notify(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "size changed") {
		t.Errorf("expected 'size changed' in output, got: %s", out)
	}
	if !strings.Contains(out, "/etc/app/config.yaml") {
		t.Errorf("expected path in output, got: %s", out)
	}
	if !strings.Contains(out, string(alert.LevelWarn)) {
		t.Errorf("expected WARN level in output, got: %s", out)
	}
}

func TestNotify_FileMissing(t *testing.T) {
	var buf bytes.Buffer
	n := alert.New(&buf)

	ev := watcher.DriftEvent{
		Path:     "/etc/app/secret.conf",
		Snapshot: makeSnapshot(),
		Current:  nil,
	}

	if err := n.Notify(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "file missing") {
		t.Errorf("expected 'file missing' in output, got: %s", out)
	}
}

func TestNotify_ModeChanged(t *testing.T) {
	var buf bytes.Buffer
	n := alert.New(&buf)

	current := makeSnapshot()
	current.Mode = fs.FileMode(0o777)

	ev := watcher.DriftEvent{
		Path:     "/etc/app/run.sh",
		Snapshot: makeSnapshot(),
		Current:  &current,
	}

	if err := n.Notify(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "mode changed") {
		t.Errorf("expected 'mode changed' in output, got: %s", out)
	}
}

func TestNew_DefaultsToStderr(t *testing.T) {
	// Ensure New(nil) doesn't panic and returns a usable Notifier.
	n := alert.New(nil)
	if n == nil {
		t.Fatal("expected non-nil Notifier")
	}
}

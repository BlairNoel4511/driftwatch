package pipeline_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/alert"
	"github.com/yourorg/driftwatch/internal/checkpoint"
	"github.com/yourorg/driftwatch/internal/filter"
	"github.com/yourorg/driftwatch/internal/history"
	"github.com/yourorg/driftwatch/internal/metrics"
	"github.com/yourorg/driftwatch/internal/notifier"
	"github.com/yourorg/driftwatch/internal/pipeline"
	"github.com/yourorg/driftwatch/internal/watcher"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return p
}

func buildPipeline(t *testing.T, paths []string) (*pipeline.Pipeline, string) {
	t.Helper()
	dir := t.TempDir()

	w, err := watcher.New(paths)
	if err != nil {
		t.Fatalf("watcher.New: %v", err)
	}
	f, _ := filter.New(nil, nil)
	a := alert.New(os.Stderr)
	h := history.New(32)
	n := notifier.New(notifier.NewAlertSink(a), h)
	cp := checkpoint.New(dir)
	m := metrics.New()

	p := pipeline.New(pipeline.Config{
		Watcher:    w,
		Filter:     f,
		Alerter:    a,
		Notifier:   n,
		Checkpoint: cp,
		Metrics:    m,
	})
	return p, dir
}

func TestRun_NoDrift(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "cfg.yaml", "key: value")

	pl, _ := buildPipeline(t, []string{path})
	ctx := context.Background()

	// First run seeds the checkpoint.
	if _, err := pl.Run(ctx); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	// Second run with no changes should detect zero drift.
	count, err := pl.Run(ctx)
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 drift events, got %d", count)
	}
}

func TestRun_DetectsDrift(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "cfg.yaml", "key: original")

	pl, _ := buildPipeline(t, []string{path})
	ctx := context.Background()

	if _, err := pl.Run(ctx); err != nil {
		t.Fatalf("seed Run: %v", err)
	}

	// Mutate the file so drift is detectable.
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(path, []byte("key: changed"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	count, err := pl.Run(ctx)
	if err != nil {
		t.Fatalf("drift Run: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 drift event, got %d", count)
	}
}

func TestRun_RecordsMetrics(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "app.conf", "x=1")

	pl, _ := buildPipeline(t, []string{path})
	ctx := context.Background()

	if _, err := pl.Run(ctx); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Metrics are internal; we only verify Run does not error out.
	_ = pl
}

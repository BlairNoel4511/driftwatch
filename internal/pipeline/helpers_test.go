package pipeline_test

import (
	"os"
	"testing"

	"github.com/yourorg/driftwatch/internal/alert"
	"github.com/yourorg/driftwatch/internal/checkpoint"
	"github.com/yourorg/driftwatch/internal/filter"
	"github.com/yourorg/driftwatch/internal/history"
	"github.com/yourorg/driftwatch/internal/metrics"
	"github.com/yourorg/driftwatch/internal/notifier"
	"github.com/yourorg/driftwatch/internal/pipeline"
	"github.com/yourorg/driftwatch/internal/watcher"
)

// buildPipelineWithExclude constructs a Pipeline whose filter excludes the
// supplied glob patterns.
func buildPipelineWithExclude(t *testing.T, paths, excludes []string) (*pipeline.Pipeline, string) {
	t.Helper()
	dir := t.TempDir()

	w, err := watcher.New(paths)
	if err != nil {
		t.Fatalf("watcher.New: %v", err)
	}
	f, err := filter.New(nil, excludes)
	if err != nil {
		t.Fatalf("filter.New: %v", err)
	}
	a := alert.New(os.Stderr)
	h := history.New(32)
	n := notifier.New(notifier.NewAlertSink(a), h)
	cp := checkpoint.New(dir)
	m := metrics.New()

	pl := pipeline.New(pipeline.Config{
		Watcher:    w,
		Filter:     f,
		Alerter:    a,
		Notifier:   n,
		Checkpoint: cp,
		Metrics:    m,
	})
	return pl, dir
}

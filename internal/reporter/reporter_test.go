package reporter_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/driftwatch/driftwatch/internal/reporter"
	"github.com/driftwatch/driftwatch/internal/watcher"
)

func makeSnapshots() (watcher.Snapshot, watcher.Snapshot) {
	baseline := watcher.Snapshot{
		Path:    "/etc/app/config.yaml",
		Size:    1024,
		Mode:    os.FileMode(0644),
		ModTime: time.Now().Add(-10 * time.Minute),
		Exists:  true,
	}
	current := watcher.Snapshot{
		Path:    "/etc/app/config.yaml",
		Size:    2048,
		Mode:    os.FileMode(0644),
		ModTime: time.Now(),
		Exists:  true,
	}
	return baseline, current
}

func TestReport_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf, reporter.FormatText)
	baseline, current := makeSnapshots()

	if err := r.Report("/etc/app/config.yaml", "size changed", baseline, current); err != nil {
		t.Fatalf("Report() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "DRIFT DETECTED") {
		t.Errorf("expected DRIFT DETECTED in output, got: %s", out)
	}
	if !strings.Contains(out, "size changed") {
		t.Errorf("expected reason in output, got: %s", out)
	}
	if !strings.Contains(out, "/etc/app/config.yaml") {
		t.Errorf("expected path in output, got: %s", out)
	}
}

func TestReport_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf, reporter.FormatJSON)
	baseline, current := makeSnapshots()

	if err := r.Report("/etc/app/config.yaml", "size changed", baseline, current); err != nil {
		t.Fatalf("Report() error: %v", err)
	}

	var rpt reporter.DriftReport
	if err := json.NewDecoder(&buf).Decode(&rpt); err != nil {
		t.Fatalf("failed to decode JSON output: %v", err)
	}
	if rpt.Path != "/etc/app/config.yaml" {
		t.Errorf("expected path /etc/app/config.yaml, got %s", rpt.Path)
	}
	if rpt.Reason != "size changed" {
		t.Errorf("expected reason 'size changed', got %s", rpt.Reason)
	}
	if rpt.Baseline.Size != 1024 {
		t.Errorf("expected baseline size 1024, got %d", rpt.Baseline.Size)
	}
	if rpt.Current.Size != 2048 {
		t.Errorf("expected current size 2048, got %d", rpt.Current.Size)
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	r := reporter.New(nil, "")
	if r == nil {
		t.Fatal("expected non-nil reporter")
	}
}

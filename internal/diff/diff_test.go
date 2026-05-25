package diff_test

import (
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/user/driftwatch/internal/diff"
	"github.com/user/driftwatch/internal/watcher"
)

var (
	t0 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 = t0.Add(time.Hour)
)

func baseSnap() *watcher.Snapshot {
	return &watcher.Snapshot{
		Path:    "/etc/app.conf",
		Size:    512,
		Mode:    0o644,
		ModTime: t0,
	}
}

func TestCompare_NoDrift(t *testing.T) {
	a, b := baseSnap(), baseSnap()
	res := diff.Compare("/etc/app.conf", a, b)
	if res.HasDrift() {
		t.Fatalf("expected no drift, got %s", res.Summary())
	}
}

func TestCompare_SizeChanged(t *testing.T) {
	a, b := baseSnap(), baseSnap()
	b.Size = 1024
	res := diff.Compare("/etc/app.conf", a, b)
	if !res.HasDrift() {
		t.Fatal("expected drift on size change")
	}
	if res.Changes[0].Field != diff.FieldSize {
		t.Errorf("expected field %q, got %q", diff.FieldSize, res.Changes[0].Field)
	}
}

func TestCompare_ModeChanged(t *testing.T) {
	a, b := baseSnap(), baseSnap()
	b.Mode = fs.FileMode(0o600)
	res := diff.Compare("/etc/app.conf", a, b)
	if !res.HasDrift() {
		t.Fatal("expected drift on mode change")
	}
	if res.Changes[0].Field != diff.FieldMode {
		t.Errorf("expected field %q, got %q", diff.FieldMode, res.Changes[0].Field)
	}
}

func TestCompare_ModTimeChanged(t *testing.T) {
	a, b := baseSnap(), baseSnap()
	b.ModTime = t1
	res := diff.Compare("/etc/app.conf", a, b)
	if !res.HasDrift() {
		t.Fatal("expected drift on mod_time change")
	}
	if res.Changes[0].Field != diff.FieldModTime {
		t.Errorf("expected field %q, got %q", diff.FieldModTime, res.Changes[0].Field)
	}
}

func TestCompare_CurrentNil_FileMissing(t *testing.T) {
	res := diff.Compare("/etc/app.conf", baseSnap(), nil)
	if !res.HasDrift() {
		t.Fatal("expected drift when current is nil")
	}
	if res.Changes[0].Field != diff.FieldMissing {
		t.Errorf("expected field %q, got %q", diff.FieldMissing, res.Changes[0].Field)
	}
}

func TestCompare_BaselineNil_FileAppeared(t *testing.T) {
	res := diff.Compare("/etc/app.conf", nil, baseSnap())
	if !res.HasDrift() {
		t.Fatal("expected drift when baseline is nil")
	}
	if res.Changes[0].NewValue != "present" {
		t.Errorf("unexpected new value: %s", res.Changes[0].NewValue)
	}
}

func TestResult_Summary_ContainsPath(t *testing.T) {
	a, b := baseSnap(), baseSnap()
	b.Size = 999
	res := diff.Compare("/etc/app.conf", a, b)
	if !strings.Contains(res.Summary(), "/etc/app.conf") {
		t.Errorf("summary missing path: %s", res.Summary())
	}
}

// TestCompare_MultipleFieldsChanged verifies that all changed fields are
// reported when more than one attribute drifts in a single comparison.
func TestCompare_MultipleFieldsChanged(t *testing.T) {
	a, b := baseSnap(), baseSnap()
	b.Size = 2048
	b.Mode = fs.FileMode(0o600)
	b.ModTime = t1
	res := diff.Compare("/etc/app.conf", a, b)
	if !res.HasDrift() {
		t.Fatal("expected drift on multiple field changes")
	}
	if len(res.Changes) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(res.Changes))
	}
	wantFields := []string{diff.FieldSize, diff.FieldMode, diff.FieldModTime}
	for i, want := range wantFields {
		if res.Changes[i].Field != want {
			t.Errorf("change[%d]: expected field %q, got %q", i, want, res.Changes[i].Field)
		}
	}
}

package envelope_test

import (
	"testing"
	"time"

	"github.com/driftwatch/driftwatch/internal/envelope"
	"github.com/driftwatch/driftwatch/internal/watcher"
)

func makeEvent(path string) watcher.Event {
	return watcher.Event{Path: path}
}

func TestNew_SetsFields(t *testing.T) {
	ev := makeEvent("/etc/app.conf")
	env := envelope.New("id-1", "drift.warning", envelope.SeverityWarning, ev)

	if env.ID != "id-1" {
		t.Errorf("expected ID id-1, got %s", env.ID)
	}
	if env.Topic != "drift.warning" {
		t.Errorf("expected topic drift.warning, got %s", env.Topic)
	}
	if env.Severity != envelope.SeverityWarning {
		t.Errorf("expected SeverityWarning, got %v", env.Severity)
	}
	if env.Event.Path != "/etc/app.conf" {
		t.Errorf("unexpected event path: %s", env.Event.Path)
	}
	if env.Attempt != 0 {
		t.Errorf("expected Attempt 0, got %d", env.Attempt)
	}
	if env.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if time.Since(env.CreatedAt) > 2*time.Second {
		t.Error("CreatedAt appears stale")
	}
}

func TestNew_PanicsOnEmptyID(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on empty id")
		}
	}()
	envelope.New("", "drift.info", envelope.SeverityInfo, makeEvent("/a"))
}

func TestNew_PanicsOnEmptyTopic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on empty topic")
		}
	}()
	envelope.New("id-2", "", envelope.SeverityInfo, makeEvent("/a"))
}

func TestRetry_IncrementsAttempt(t *testing.T) {
	env := envelope.New("id-3", "drift.critical", envelope.SeverityCritical, makeEvent("/b"))

	r1 := env.Retry()
	if r1.Attempt != 1 {
		t.Errorf("expected Attempt 1 after first retry, got %d", r1.Attempt)
	}

	r2 := r1.Retry()
	if r2.Attempt != 2 {
		t.Errorf("expected Attempt 2 after second retry, got %d", r2.Attempt)
	}

	// original must be unchanged
	if env.Attempt != 0 {
		t.Errorf("original Attempt should remain 0, got %d", env.Attempt)
	}
}

func TestSeverity_String(t *testing.T) {
	cases := []struct {
		s    envelope.Severity
		want string
	}{
		{envelope.SeverityInfo, "info"},
		{envelope.SeverityWarning, "warning"},
		{envelope.SeverityCritical, "critical"},
		{envelope.Severity(99), "severity(99)"},
	}
	for _, tc := range cases {
		if got := tc.s.String(); got != tc.want {
			t.Errorf("Severity(%d).String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}

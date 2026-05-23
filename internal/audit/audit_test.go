package audit_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"driftwatch/internal/audit"
)

func TestNew_DefaultsToStderr(t *testing.T) {
	// Ensure New(nil) does not panic.
	l := audit.New(nil)
	if l == nil {
		t.Fatal("expected non-nil Log")
	}
}

func TestRecord_WritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)

	if err := l.Record(audit.KindDetected, "/etc/hosts", "size changed"); err != nil {
		t.Fatalf("Record: %v", err)
	}

	var entry audit.Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry.Kind != audit.KindDetected {
		t.Errorf("kind: got %q, want %q", entry.Kind, audit.KindDetected)
	}
	if entry.Path != "/etc/hosts" {
		t.Errorf("path: got %q, want /etc/hosts", entry.Path)
	}
	if entry.Detail != "size changed" {
		t.Errorf("detail: got %q, want \"size changed\"", entry.Detail)
	}
	if entry.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
	if entry.Timestamp.Location() != time.UTC {
		t.Error("timestamp should be UTC")
	}
}

func TestDetected_Convenience(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)
	if err := l.Detected("/etc/passwd", "mode changed"); err != nil {
		t.Fatalf("Detected: %v", err)
	}
	if !strings.Contains(buf.String(), `"detected"`) {
		t.Errorf("expected kind detected in output: %s", buf.String())
	}
}

func TestAcknowledged_Convenience(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)
	if err := l.Acknowledged("/etc/passwd", "operator ack"); err != nil {
		t.Fatalf("Acknowledged: %v", err)
	}
	if !strings.Contains(buf.String(), `"acknowledged"`) {
		t.Errorf("expected kind acknowledged in output: %s", buf.String())
	}
}

func TestResolved_Convenience(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)
	if err := l.Resolved("/etc/ssh/sshd_config", "drift cleared"); err != nil {
		t.Fatalf("Resolved: %v", err)
	}
	if !strings.Contains(buf.String(), `"resolved"`) {
		t.Errorf("expected kind resolved in output: %s", buf.String())
	}
}

func TestRecord_MultipleEntriesEachOnOwnLine(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)
	_ = l.Detected("/a", "")
	_ = l.Resolved("/a", "")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), buf.String())
	}
	for i, line := range lines {
		var e audit.Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Errorf("line %d unmarshal: %v", i, err)
		}
	}
}

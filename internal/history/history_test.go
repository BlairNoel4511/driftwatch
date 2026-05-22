package history_test

import (
	"fmt"
	"testing"

	"github.com/yourorg/driftwatch/internal/history"
)

func TestNew_DefaultCapacity(t *testing.T) {
	h := history.New(0)
	if h == nil {
		t.Fatal("expected non-nil History")
	}
}

func TestRecord_And_All(t *testing.T) {
	h := history.New(10)
	h.Record("/etc/hosts", "size changed")
	h.Record("/etc/passwd", "mode changed")

	events := h.All()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Path != "/etc/hosts" {
		t.Errorf("expected /etc/hosts, got %s", events[0].Path)
	}
	if events[1].Reason != "mode changed" {
		t.Errorf("unexpected reason: %s", events[1].Reason)
	}
}

func TestRecord_Evicts_OldestWhenFull(t *testing.T) {
	h := history.New(3)
	for i := 0; i < 5; i++ {
		h.Record(fmt.Sprintf("/file/%d", i), "drift")
	}

	events := h.All()
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	// oldest surviving entry should be index 2
	if events[0].Path != "/file/2" {
		t.Errorf("expected /file/2 as oldest, got %s", events[0].Path)
	}
}

func TestLen(t *testing.T) {
	h := history.New(10)
	if h.Len() != 0 {
		t.Fatalf("expected 0, got %d", h.Len())
	}
	h.Record("/a", "x")
	h.Record("/b", "y")
	if h.Len() != 2 {
		t.Fatalf("expected 2, got %d", h.Len())
	}
}

func TestClear(t *testing.T) {
	h := history.New(10)
	h.Record("/etc/hosts", "size changed")
	h.Clear()
	if h.Len() != 0 {
		t.Fatalf("expected 0 after clear, got %d", h.Len())
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	h := history.New(10)
	h.Record("/etc/hosts", "size changed")

	events := h.All()
	events[0].Path = "mutated"

	again := h.All()
	if again[0].Path == "mutated" {
		t.Error("All() should return a copy, not a reference to internal slice")
	}
}

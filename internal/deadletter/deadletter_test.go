package deadletter_test

import (
	"testing"
	"time"

	"github.com/driftwatch/driftwatch/internal/deadletter"
	"github.com/driftwatch/driftwatch/internal/watcher"
)

func makeEvent(path string) watcher.DriftEvent {
	return watcher.DriftEvent{
		Path:      path,
		DetectedAt: time.Now(),
	}
}

func TestNew_PanicsOnZeroCapacity(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero capacity")
		}
	}()
	deadletter.New(0)
}

func TestPush_And_All(t *testing.T) {
	q := deadletter.New(10)
	q.Push(makeEvent("/etc/hosts"), "timeout", 3)
	q.Push(makeEvent("/etc/resolv.conf"), "refused", 1)

	entries := q.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Event.Path != "/etc/hosts" {
		t.Errorf("unexpected path: %s", entries[0].Event.Path)
	}
	if entries[0].Reason != "timeout" {
		t.Errorf("unexpected reason: %s", entries[0].Reason)
	}
	if entries[0].Attempts != 3 {
		t.Errorf("unexpected attempts: %d", entries[0].Attempts)
	}
}

func TestPush_EvictsOldestWhenFull(t *testing.T) {
	q := deadletter.New(3)
	q.Push(makeEvent("/a"), "err", 1)
	q.Push(makeEvent("/b"), "err", 1)
	q.Push(makeEvent("/c"), "err", 1)
	q.Push(makeEvent("/d"), "err", 1) // should evict /a

	entries := q.All()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(entries))
	}
	if entries[0].Event.Path != "/b" {
		t.Errorf("expected oldest to be evicted; first entry path = %s", entries[0].Event.Path)
	}
	if entries[2].Event.Path != "/d" {
		t.Errorf("expected /d as last entry; got %s", entries[2].Event.Path)
	}
}

func TestLen(t *testing.T) {
	q := deadletter.New(5)
	if q.Len() != 0 {
		t.Fatalf("expected empty queue")
	}
	q.Push(makeEvent("/x"), "err", 1)
	if q.Len() != 1 {
		t.Fatalf("expected len 1, got %d", q.Len())
	}
}

func TestClear(t *testing.T) {
	q := deadletter.New(5)
	q.Push(makeEvent("/x"), "err", 1)
	q.Push(makeEvent("/y"), "err", 2)
	q.Clear()
	if q.Len() != 0 {
		t.Fatalf("expected empty queue after Clear, got %d", q.Len())
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	q := deadletter.New(5)
	q.Push(makeEvent("/z"), "err", 1)

	a := q.All()
	a[0].Reason = "mutated"

	b := q.All()
	if b[0].Reason == "mutated" {
		t.Error("All() should return an independent copy")
	}
}

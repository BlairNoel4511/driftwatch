package correlator_test

import (
	"path/filepath"
	"testing"

	"github.com/driftwatch/driftwatch/internal/correlator"
	"github.com/driftwatch/driftwatch/internal/watcher"
)

func dirKey(e watcher.Event) string { return filepath.Dir(e.Path) }

func TestNew_PanicsOnNilKeyFunc(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil keyFunc")
		}
	}()
	correlator.New(nil)
}

func TestRecord_GroupsByKey(t *testing.T) {
	c := correlator.New(dirKey)
	c.Record(watcher.Event{Path: "/etc/app/a.conf"})
	c.Record(watcher.Event{Path: "/etc/app/b.conf"})
	c.Record(watcher.Event{Path: "/etc/other/c.conf"})

	if got := c.Len(); got != 2 {
		t.Fatalf("expected 2 groups, got %d", got)
	}
}

func TestRecord_AccumulatesEventsInGroup(t *testing.T) {
	c := correlator.New(dirKey)
	c.Record(watcher.Event{Path: "/etc/app/a.conf"})
	c.Record(watcher.Event{Path: "/etc/app/b.conf"})

	groups := c.Groups()
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if got := len(groups[0].Events); got != 2 {
		t.Fatalf("expected 2 events in group, got %d", got)
	}
}

func TestGroups_ReturnsCopy(t *testing.T) {
	c := correlator.New(dirKey)
	c.Record(watcher.Event{Path: "/etc/app/a.conf"})

	g1 := c.Groups()
	g1[0].Events = nil // mutate the copy

	g2 := c.Groups()
	if len(g2[0].Events) == 0 {
		t.Fatal("Groups() should return independent copies")
	}
}

func TestClear_RemovesAllGroups(t *testing.T) {
	c := correlator.New(dirKey)
	c.Record(watcher.Event{Path: "/etc/app/a.conf"})
	c.Clear()

	if got := c.Len(); got != 0 {
		t.Fatalf("expected 0 groups after Clear, got %d", got)
	}
}

func TestRecord_TimestampsPopulated(t *testing.T) {
	c := correlator.New(dirKey)
	c.Record(watcher.Event{Path: "/etc/app/a.conf"})
	c.Record(watcher.Event{Path: "/etc/app/b.conf"})

	groups := c.Groups()
	g := groups[0]
	if g.First.IsZero() {
		t.Error("First timestamp should not be zero")
	}
	if g.Last.IsZero() {
		t.Error("Last timestamp should not be zero")
	}
	if g.Last.Before(g.First) {
		t.Error("Last should be >= First")
	}
}

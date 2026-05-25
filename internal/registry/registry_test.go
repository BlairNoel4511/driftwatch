package registry

import (
	"testing"
)

func TestRegister_And_Get_RoundTrip(t *testing.T) {
	r := New()
	e := Entry{Path: "/etc/app.conf", Enabled: true, Tags: []string{"prod"}}
	if err := r.Register(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := r.Get("/etc/app.conf")
	if !ok {
		t.Fatal("expected entry to be present")
	}
	if got.Path != e.Path || got.Enabled != e.Enabled {
		t.Fatalf("got %+v, want %+v", got, e)
	}
}

func TestRegister_EmptyPathReturnsError(t *testing.T) {
	r := New()
	if err := r.Register(Entry{}); err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}

func TestGet_ReturnsFalseWhenUnset(t *testing.T) {
	r := New()
	_, ok := r.Get("/nonexistent")
	if ok {
		t.Fatal("expected false for unregistered path")
	}
}

func TestDeregister_RemovesEntry(t *testing.T) {
	r := New()
	_ = r.Register(Entry{Path: "/tmp/x"})
	r.Deregister("/tmp/x")
	_, ok := r.Get("/tmp/x")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestDeregister_NoopOnMissingPath(t *testing.T) {
	r := New()
	// should not panic
	r.Deregister("/does/not/exist")
}

func TestAll_ReturnsAllEntries(t *testing.T) {
	r := New()
	paths := []string{"/a", "/b", "/c"}
	for _, p := range paths {
		_ = r.Register(Entry{Path: p, Enabled: true})
	}
	all := r.All()
	if len(all) != len(paths) {
		t.Fatalf("got %d entries, want %d", len(all), len(paths))
	}
}

func TestAll_ReturnsEmptySliceWhenNoEntries(t *testing.T) {
	r := New()
	all := r.All()
	if all == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
	if len(all) != 0 {
		t.Fatalf("expected empty slice, got %d entries", len(all))
	}
}

func TestLen_ReflectsRegistrations(t *testing.T) {
	r := New()
	if r.Len() != 0 {
		t.Fatalf("expected 0, got %d", r.Len())
	}
	_ = r.Register(Entry{Path: "/x"})
	_ = r.Register(Entry{Path: "/y"})
	if r.Len() != 2 {
		t.Fatalf("expected 2, got %d", r.Len())
	}
	r.Deregister("/x")
	if r.Len() != 1 {
		t.Fatalf("expected 1 after deregister, got %d", r.Len())
	}
}

func TestRegister_ReplacesExistingEntry(t *testing.T) {
	r := New()
	_ = r.Register(Entry{Path: "/etc/app.conf", Enabled: true})
	_ = r.Register(Entry{Path: "/etc/app.conf", Enabled: false, Tags: []string{"updated"}})
	got, _ := r.Get("/etc/app.conf")
	if got.Enabled {
		t.Fatal("expected Enabled to be false after replacement")
	}
	if len(got.Tags) != 1 || got.Tags[0] != "updated" {
		t.Fatalf("unexpected tags: %v", got.Tags)
	}
}

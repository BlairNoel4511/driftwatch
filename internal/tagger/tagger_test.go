package tagger_test

import (
	"sort"
	"testing"

	"driftwatch/internal/tagger"
)

func TestGet_ReturnsNilWhenUnset(t *testing.T) {
	tr := tagger.New()
	if got := tr.Get("/etc/hosts"); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestSet_And_Get_RoundTrip(t *testing.T) {
	tr := tagger.New()
	tr.Set("/etc/hosts", []string{"network", "critical"})
	got := tr.Get("/etc/hosts")
	if len(got) != 2 || got[0] != "network" || got[1] != "critical" {
		t.Fatalf("unexpected tags: %v", got)
	}
}

func TestSet_ReplacesExistingTags(t *testing.T) {
	tr := tagger.New()
	tr.Set("/etc/hosts", []string{"old"})
	tr.Set("/etc/hosts", []string{"new"})
	got := tr.Get("/etc/hosts")
	if len(got) != 1 || got[0] != "new" {
		t.Fatalf("expected [new], got %v", got)
	}
}

func TestAdd_DeduplicatesTags(t *testing.T) {
	tr := tagger.New()
	tr.Set("/etc/hosts", []string{"a", "b"})
	tr.Add("/etc/hosts", "b", "c")
	got := tr.Get("/etc/hosts")
	sort.Strings(got)
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("expected [a b c], got %v", got)
	}
}

func TestAdd_WorksOnUnsetPath(t *testing.T) {
	tr := tagger.New()
	tr.Add("/etc/resolv.conf", "dns")
	got := tr.Get("/etc/resolv.conf")
	if len(got) != 1 || got[0] != "dns" {
		t.Fatalf("expected [dns], got %v", got)
	}
}

func TestRemove_ClearsTags(t *testing.T) {
	tr := tagger.New()
	tr.Set("/etc/hosts", []string{"x"})
	tr.Remove("/etc/hosts")
	if got := tr.Get("/etc/hosts"); got != nil {
		t.Fatalf("expected nil after remove, got %v", got)
	}
}

func TestPaths_ReturnsTaggedPaths(t *testing.T) {
	tr := tagger.New()
	tr.Set("/a", []string{"t1"})
	tr.Set("/b", []string{"t2"})
	paths := tr.Paths()
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	tr := tagger.New()
	tr.Set("/etc/hosts", []string{"immutable"})
	got := tr.Get("/etc/hosts")
	got[0] = "mutated"
	if tr.Get("/etc/hosts")[0] != "immutable" {
		t.Fatal("Get should return a copy, not a reference")
	}
}

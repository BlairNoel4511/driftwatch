package labelset_test

import (
	"testing"

	"driftwatch/internal/labelset"
)

const path = "/etc/app/config.yaml"

func TestGet_ReturnsFalseWhenUnset(t *testing.T) {
	ls := labelset.New()
	_, ok := ls.Get(path)
	if ok {
		t.Fatal("expected false for unset path")
	}
}

func TestPut_And_Get_RoundTrip(t *testing.T) {
	ls := labelset.New()
	if err := ls.Put(path, "env", "prod"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	labels, ok := ls.Get(path)
	if !ok {
		t.Fatal("expected labels to exist")
	}
	if labels["env"] != "prod" {
		t.Fatalf("got %q, want %q", labels["env"], "prod")
	}
}

func TestPut_EmptyKeyReturnsError(t *testing.T) {
	ls := labelset.New()
	if err := ls.Put(path, "", "value"); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestSet_ReplacesAllLabels(t *testing.T) {
	ls := labelset.New()
	_ = ls.Put(path, "old", "value")
	ls.Set(path, map[string]string{"new": "label"})
	labels, _ := ls.Get(path)
	if _, exists := labels["old"]; exists {
		t.Fatal("old label should have been replaced")
	}
	if labels["new"] != "label" {
		t.Fatal("new label not found")
	}
}

func TestSet_NilClearsLabels(t *testing.T) {
	ls := labelset.New()
	_ = ls.Put(path, "k", "v")
	ls.Set(path, nil)
	_, ok := ls.Get(path)
	if ok {
		t.Fatal("expected labels to be cleared")
	}
}

func TestDelete_RemovesSingleKey(t *testing.T) {
	ls := labelset.New()
	_ = ls.Put(path, "a", "1")
	_ = ls.Put(path, "b", "2")
	ls.Delete(path, "a")
	labels, _ := ls.Get(path)
	if _, exists := labels["a"]; exists {
		t.Fatal("key 'a' should have been deleted")
	}
	if labels["b"] != "2" {
		t.Fatal("key 'b' should still exist")
	}
}

func TestDelete_RemovesPathWhenEmpty(t *testing.T) {
	ls := labelset.New()
	_ = ls.Put(path, "only", "key")
	ls.Delete(path, "only")
	_, ok := ls.Get(path)
	if ok {
		t.Fatal("path should be removed when all labels deleted")
	}
}

func TestPaths_ReturnsLabelledPaths(t *testing.T) {
	ls := labelset.New()
	_ = ls.Put("/a", "k", "v")
	_ = ls.Put("/b", "k", "v")
	paths := ls.Paths()
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	ls := labelset.New()
	_ = ls.Put(path, "env", "staging")
	labels, _ := ls.Get(path)
	labels["env"] = "mutated"
	labels2, _ := ls.Get(path)
	if labels2["env"] != "staging" {
		t.Fatal("Get should return an independent copy")
	}
}

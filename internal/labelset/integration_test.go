package labelset_test

import (
	"sync"
	"testing"

	"driftwatch/internal/labelset"
)

func TestLabelSet_ConcurrentAccess(t *testing.T) {
	ls := labelset.New()
	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			p := "/etc/file"
			_ = ls.Put(p, "worker", "yes")
			_, _ = ls.Get(p)
			ls.Delete(p, "worker")
		}(i)
	}
	wg.Wait()
}

func TestLabelSet_MultiplePaths(t *testing.T) {
	ls := labelset.New()
	paths := []string{"/etc/a", "/etc/b", "/etc/c"}
	for _, p := range paths {
		_ = ls.Put(p, "managed", "true")
	}
	got := ls.Paths()
	if len(got) != len(paths) {
		t.Fatalf("expected %d paths, got %d", len(paths), len(got))
	}
	// Clear one path via Set(nil) and verify count drops.
	ls.Set("/etc/a", nil)
	got = ls.Paths()
	if len(got) != 2 {
		t.Fatalf("expected 2 paths after clear, got %d", len(got))
	}
}

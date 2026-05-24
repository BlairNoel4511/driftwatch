package snapshot_test

import (
	"sync"
	"testing"

	"github.com/driftwatch/driftwatch/internal/snapshot"
)

func TestStore_ConcurrentAccess(t *testing.T) {
	s := snapshot.New()
	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(n int) {
			defer wg.Done()
			path := "/concurrent/path"
			_ = s.Set(makeEntry(path))
			_, _ = s.Get(path)
			_ = s.All()
		}(i)
	}
	wg.Wait()
	if s.Len() != 1 {
		t.Errorf("expected 1 unique path, got %d", s.Len())
	}
}

func TestStore_MultiplePaths(t *testing.T) {
	s := snapshot.New()
	paths := []string{"/etc/hosts", "/etc/resolv.conf", "/etc/nsswitch.conf"}
	for _, p := range paths {
		if err := s.Set(makeEntry(p)); err != nil {
			t.Fatalf("Set(%q): %v", p, err)
		}
	}
	for _, p := range paths {
		e, ok := s.Get(p)
		if !ok {
			t.Errorf("Get(%q): not found", p)
		}
		if e.Path != p {
			t.Errorf("Get(%q).Path = %q", p, e.Path)
		}
	}
	if s.Len() != len(paths) {
		t.Errorf("Len: got %d, want %d", s.Len(), len(paths))
	}
}

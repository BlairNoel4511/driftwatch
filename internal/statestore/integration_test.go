package statestore_test

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"driftwatch/internal/statestore"
)

func TestStore_ConcurrentAccess(t *testing.T) {
	st, err := statestore.New(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(n int) {
			defer wg.Done()
			path := fmt.Sprintf("/etc/file%d", n)
			_ = st.Set(makeEntry(path))
			_, _ = st.Get(path)
		}(i)
	}
	wg.Wait()

	all := st.All()
	if len(all) != workers {
		t.Errorf("got %d entries, want %d", len(all), workers)
	}
}

func TestStore_DeleteDoesNotAffectOtherPaths(t *testing.T) {
	st, _ := statestore.New(filepath.Join(t.TempDir(), "state.json"))

	for i := 0; i < 5; i++ {
		_ = st.Set(makeEntry(fmt.Sprintf("/etc/file%d", i)))
	}

	_ = st.Delete("/etc/file2")

	all := st.All()
	if len(all) != 4 {
		t.Errorf("got %d entries after delete, want 4", len(all))
	}
	for _, e := range all {
		if e.Path == "/etc/file2" {
			t.Error("deleted entry still present")
		}
	}
}

package statestore_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"driftwatch/internal/statestore"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "state.json")
}

func makeEntry(path string) statestore.Entry {
	return statestore.Entry{
		Path:       path,
		Size:       1024,
		Mode:       0o644,
		ModTime:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Checksum:   "abc123",
		RecordedAt: time.Now(),
	}
}

func TestNew_CreatesFile(t *testing.T) {
	p := tempPath(t)
	_, err := statestore.New(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSet_And_Get_RoundTrip(t *testing.T) {
	st, _ := statestore.New(tempPath(t))
	e := makeEntry("/etc/hosts")
	if err := st.Set(e); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok := st.Get("/etc/hosts")
	if !ok {
		t.Fatal("expected entry to be present")
	}
	if got.Size != e.Size || got.Checksum != e.Checksum {
		t.Errorf("got %+v, want %+v", got, e)
	}
}

func TestGet_ReturnsFalseWhenUnset(t *testing.T) {
	st, _ := statestore.New(tempPath(t))
	_, ok := st.Get("/nonexistent")
	if ok {
		t.Fatal("expected false for missing path")
	}
}

func TestSet_EmptyPathReturnsError(t *testing.T) {
	st, _ := statestore.New(tempPath(t))
	if err := st.Set(statestore.Entry{}); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	st, _ := statestore.New(tempPath(t))
	_ = st.Set(makeEntry("/etc/hosts"))
	if err := st.Delete("/etc/hosts"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, ok := st.Get("/etc/hosts")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestAll_ReturnsAllEntries(t *testing.T) {
	st, _ := statestore.New(tempPath(t))
	paths := []string{"/etc/hosts", "/etc/passwd", "/etc/resolv.conf"}
	for _, p := range paths {
		_ = st.Set(makeEntry(p))
	}
	all := st.All()
	if len(all) != len(paths) {
		t.Errorf("got %d entries, want %d", len(all), len(paths))
	}
}

func TestPersistence_ReloadsFromDisk(t *testing.T) {
	p := tempPath(t)
	st1, _ := statestore.New(p)
	_ = st1.Set(makeEntry("/etc/hosts"))

	st2, err := statestore.New(p)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	_, ok := st2.Get("/etc/hosts")
	if !ok {
		t.Fatal("expected entry to survive reload")
	}
}

func TestNew_CreatesParentDirectory(t *testing.T) {
	base := t.TempDir()
	p := filepath.Join(base, "sub", "dir", "state.json")
	_, err := statestore.New(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(p)); err != nil {
		t.Errorf("parent directory not created: %v", err)
	}
}

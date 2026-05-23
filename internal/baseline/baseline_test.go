package baseline_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"driftwatch/internal/baseline"
)

func tempFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "baseline.json")
}

func makeEntry(path string) baseline.Entry {
	return baseline.Entry{
		Path:    path,
		Size:    1024,
		Mode:    0o644,
		ModTime: time.Now().Truncate(time.Second),
		Hash:    "abc123",
	}
}

func TestNew_CreatesFileOnFirstSave(t *testing.T) {
	f := tempFile(t)
	s, err := baseline.New(f)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := s.Set(makeEntry("/etc/hosts")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, err := os.Stat(f); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}

func TestSet_And_Get_RoundTrip(t *testing.T) {
	s, _ := baseline.New(tempFile(t))
	entry := makeEntry("/etc/passwd")
	if err := s.Set(entry); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok := s.Get("/etc/passwd")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if got.Size != entry.Size || got.Hash != entry.Hash {
		t.Errorf("got %+v, want %+v", got, entry)
	}
}

func TestGet_ReturnsFalseWhenUnset(t *testing.T) {
	s, _ := baseline.New(tempFile(t))
	_, ok := s.Get("/not/tracked")
	if ok {
		t.Error("expected false for untracked path")
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	s, _ := baseline.New(tempFile(t))
	_ = s.Set(makeEntry("/etc/hosts"))
	if err := s.Delete("/etc/hosts"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok := s.Get("/etc/hosts"); ok {
		t.Error("expected entry to be deleted")
	}
}

func TestAll_ReturnsAllEntries(t *testing.T) {
	s, _ := baseline.New(tempFile(t))
	_ = s.Set(makeEntry("/a"))
	_ = s.Set(makeEntry("/b"))
	if got := len(s.All()); got != 2 {
		t.Errorf("All: got %d entries, want 2", got)
	}
}

func TestNew_LoadsExistingFile(t *testing.T) {
	f := tempFile(t)
	s1, _ := baseline.New(f)
	_ = s1.Set(makeEntry("/etc/hosts"))

	s2, err := baseline.New(f)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if _, ok := s2.Get("/etc/hosts"); !ok {
		t.Error("expected persisted entry to be reloaded")
	}
}

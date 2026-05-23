package checkpoint_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/driftwatch/driftwatch/internal/checkpoint"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "cp")
	return dir
}

func baseSnap(path string) checkpoint.Snapshot {
	return checkpoint.Snapshot{
		Path:    path,
		Size:    1024,
		Mode:    0644,
		ModTime: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		Missing: false,
	}
}

func TestNew_CreatesDirectory(t *testing.T) {
	dir := tempDir(t)
	_, err := checkpoint.New(dir)
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatal("expected directory to be created")
	}
}

func TestLoad_MissingCheckpoint_ReturnsFalse(t *testing.T) {
	store, _ := checkpoint.New(tempDir(t))
	_, ok, err := store.Load("/etc/hosts")
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for unsaved path")
	}
}

func TestSave_And_Load_RoundTrip(t *testing.T) {
	store, _ := checkpoint.New(tempDir(t))
	snap := baseSnap("/etc/nginx/nginx.conf")

	if err := store.Save(snap); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, ok, err := store.Load(snap.Path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true after Save")
	}
	if got.Size != snap.Size || got.Mode != snap.Mode || got.Path != snap.Path {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, snap)
	}
}

func TestDelete_RemovesCheckpoint(t *testing.T) {
	store, _ := checkpoint.New(tempDir(t))
	snap := baseSnap("/tmp/watched.conf")

	_ = store.Save(snap)
	if err := store.Delete(snap.Path); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, ok, err := store.Load(snap.Path)
	if err != nil {
		t.Fatalf("Load after Delete: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false after Delete")
	}
}

func TestDelete_NonExistent_NoError(t *testing.T) {
	store, _ := checkpoint.New(tempDir(t))
	if err := store.Delete("/no/such/file"); err != nil {
		t.Fatalf("Delete of non-existent path should not error: %v", err)
	}
}

func TestSave_OverwritesPreviousCheckpoint(t *testing.T) {
	store, _ := checkpoint.New(tempDir(t))
	path := "/etc/hosts"

	_ = store.Save(baseSnap(path))

	updated := baseSnap(path)
	updated.Size = 9999
	_ = store.Save(updated)

	got, _, _ := store.Load(path)
	if got.Size != 9999 {
		t.Errorf("expected updated size 9999, got %d", got.Size)
	}
}

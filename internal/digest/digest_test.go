package digest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/driftwatch/driftwatch/internal/digest"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return p
}

func TestSum_ReturnsConsistentHash(t *testing.T) {
	p := writeTempFile(t, "hello driftwatch")
	d := digest.New()

	h1, err := d.Sum(p)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	h2, err := d.Sum(p) // second call should hit cache
	if err != nil {
		t.Fatalf("Sum (cached): %v", err)
	}
	if h1 != h2 {
		t.Errorf("expected identical hashes, got %q and %q", h1, h2)
	}
}

func TestSum_DifferentContentDifferentHash(t *testing.T) {
	p1 := writeTempFile(t, "aaa")
	p2 := writeTempFile(t, "bbb")
	d := digest.New()

	h1, _ := d.Sum(p1)
	h2, _ := d.Sum(p2)
	if h1 == h2 {
		t.Error("expected different hashes for different content")
	}
}

func TestInvalidate_ForcesRecompute(t *testing.T) {
	p := writeTempFile(t, "original")
	d := digest.New()

	old, _ := d.Sum(p)

	// Mutate the file, then invalidate cache.
	if err := os.WriteFile(p, []byte("modified"), 0o644); err != nil {
		t.Fatal(err)
	}
	d.Invalidate(p)

	newHash, _ := d.Sum(p)
	if old == newHash {
		t.Error("expected new hash after invalidation and file change")
	}
}

func TestChanged_DetectsModification(t *testing.T) {
	p := writeTempFile(t, "v1")
	d := digest.New()

	baseline, err := d.Sum(p)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}

	if err := os.WriteFile(p, []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, current, err := d.Changed(p, baseline)
	if err != nil {
		t.Fatalf("Changed: %v", err)
	}
	if !changed {
		t.Error("expected Changed to return true after file modification")
	}
	if current == baseline {
		t.Error("expected current hash to differ from baseline")
	}
}

func TestChanged_NoDriftWhenUnchanged(t *testing.T) {
	p := writeTempFile(t, "stable")
	d := digest.New()

	baseline, _ := d.Sum(p)
	changed, current, err := d.Changed(p, baseline)
	if err != nil {
		t.Fatalf("Changed: %v", err)
	}
	if changed {
		t.Error("expected Changed to return false for unmodified file")
	}
	if current != baseline {
		t.Errorf("expected current == baseline, got %q vs %q", current, baseline)
	}
}

func TestSum_MissingFile_ReturnsError(t *testing.T) {
	d := digest.New()
	_, err := d.Sum("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

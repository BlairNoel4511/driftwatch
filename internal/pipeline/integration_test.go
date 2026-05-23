package pipeline_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestPipeline_MultiFile verifies that drift is detected across several
// monitored files in a single Run cycle.
func TestPipeline_MultiFile(t *testing.T) {
	dir := t.TempDir()
	paths := make([]string, 3)
	for i, name := range []string{"a.yaml", "b.yaml", "c.yaml"} {
		paths[i] = writeTempFile(t, dir, name, "v: original")
	}

	pl, _ := buildPipeline(t, paths)
	ctx := context.Background()

	if _, err := pl.Run(ctx); err != nil {
		t.Fatalf("seed Run: %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	// Mutate two of the three files.
	for _, p := range paths[:2] {
		if err := os.WriteFile(p, []byte("v: changed"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}

	count, err := pl.Run(ctx)
	if err != nil {
		t.Fatalf("drift Run: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 drift events, got %d", count)
	}
}

// TestPipeline_FilterExcludesPath verifies that excluded paths are not
// reported as drift even when their content changes.
func TestPipeline_FilterExcludesPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.yaml")
	if err := os.WriteFile(path, []byte("token: abc"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Exclude the file via glob.
	f, _ := buildPipelineWithExclude(t, []string{path}, []string{"*secret*"})
	ctx := context.Background()

	if _, err := f.Run(ctx); err != nil {
		t.Fatalf("seed Run: %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(path, []byte("token: xyz"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	count, err := f.Run(ctx)
	if err != nil {
		t.Fatalf("drift Run: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 drift events (excluded), got %d", count)
	}
}

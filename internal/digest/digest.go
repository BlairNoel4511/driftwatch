// Package digest computes and compares cryptographic hashes of monitored files,
// enabling driftwatch to detect content changes independently of mtime/size.
package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
)

// Digester computes SHA-256 hashes for file paths and caches the results.
type Digester struct {
	mu    sync.Mutex
	cache map[string]string
}

// New returns a new Digester with an empty cache.
func New() *Digester {
	return &Digester{
		cache: make(map[string]string),
	}
}

// Sum returns the hex-encoded SHA-256 hash of the file at path.
// The result is cached; call Invalidate to clear a stale entry.
func (d *Digester) Sum(path string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if h, ok := d.cache[path]; ok {
		return h, nil
	}

	h, err := hashFile(path)
	if err != nil {
		return "", fmt.Errorf("digest: hashing %q: %w", path, err)
	}

	d.cache[path] = h
	return h, nil
}

// Invalidate removes the cached hash for path, forcing a recompute on next Sum.
func (d *Digester) Invalidate(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.cache, path)
}

// Changed reports whether the current hash of path differs from prev.
// It always computes a fresh hash (bypassing the cache).
func (d *Digester) Changed(path, prev string) (bool, string, error) {
	current, err := hashFile(path)
	if err != nil {
		return false, "", fmt.Errorf("digest: hashing %q: %w", path, err)
	}
	d.mu.Lock()
	d.cache[path] = current
	d.mu.Unlock()
	return current != prev, current, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

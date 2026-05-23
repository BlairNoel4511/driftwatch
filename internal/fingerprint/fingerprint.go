// Package fingerprint derives a stable identity string for a monitored path
// by combining its metadata and content digest into a single opaque token.
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/driftwatch/internal/watcher"
)

// Fingerprinter computes and caches fingerprints for file snapshots.
type Fingerprinter struct {
	mu    sync.Mutex
	cache map[string]string // path -> last fingerprint
}

// New returns a ready-to-use Fingerprinter.
func New() *Fingerprinter {
	return &Fingerprinter{
		cache: make(map[string]string),
	}
}

// Compute derives a fingerprint from a Snapshot's observable fields.
// The fingerprint is a hex-encoded SHA-256 of path|size|mode|modtime|checksum.
func (f *Fingerprinter) Compute(snap watcher.Snapshot) string {
	raw := fmt.Sprintf("%s|%d|%s|%d|%s",
		snap.Path,
		snap.Size,
		snap.Mode.String(),
		snap.ModTime.UnixNano(),
		snap.Checksum,
	)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// Changed returns true when the fingerprint of snap differs from the last
// recorded fingerprint for that path, and updates the internal cache.
func (f *Fingerprinter) Changed(snap watcher.Snapshot) bool {
	next := f.Compute(snap)
	f.mu.Lock()
	defer f.mu.Unlock()
	prev, ok := f.cache[snap.Path]
	f.cache[snap.Path] = next
	if !ok {
		return false // first observation is never "changed"
	}
	return prev != next
}

// Evict removes the cached fingerprint for path, forcing the next call to
// Changed to treat the file as newly seen.
func (f *Fingerprinter) Evict(path string) {
	f.mu.Lock()
	delete(f.cache, path)
	f.mu.Unlock()
}

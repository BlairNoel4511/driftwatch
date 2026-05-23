// Package checkpoint persists watcher snapshots to disk so that driftwatch
// can resume drift detection across restarts without a false-positive burst.
package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Snapshot is a minimal, serialisable record of a file's observed state.
type Snapshot struct {
	Path    string      `json:"path"`
	Size    int64       `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime time.Time   `json:"mod_time"`
	Missing bool        `json:"missing"`
}

// Store writes and reads checkpoint files from a configurable directory.
type Store struct {
	mu  sync.RWMutex
	dir string
}

// New returns a Store that persists checkpoints under dir.
// dir is created with 0700 permissions if it does not exist.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("checkpoint: create dir %q: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// Save serialises snap to disk, keyed by snap.Path.
func (s *Store) Save(snap Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("checkpoint: marshal: %w", err)
	}
	return os.WriteFile(s.filePath(snap.Path), data, 0600)
}

// Load retrieves the last saved snapshot for path.
// Returns (zero, false, nil) when no checkpoint exists yet.
func (s *Store) Load(path string) (Snapshot, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.filePath(path))
	if os.IsNotExist(err) {
		return Snapshot{}, false, nil
	}
	if err != nil {
		return Snapshot{}, false, fmt.Errorf("checkpoint: read: %w", err)
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return Snapshot{}, false, fmt.Errorf("checkpoint: unmarshal: %w", err)
	}
	return snap, true, nil
}

// Delete removes the checkpoint for path; a missing file is not an error.
func (s *Store) Delete(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.Remove(s.filePath(path))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// filePath maps a monitored file path to a stable checkpoint filename.
func (s *Store) filePath(path string) string {
	safe := filepath.Base(path) + "-" + fmt.Sprintf("%x", len(path))
	return filepath.Join(s.dir, safe+".json")
}

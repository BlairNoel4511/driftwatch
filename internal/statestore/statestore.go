// Package statestore provides a persistent key-value store for tracking
// the last-known state of monitored paths across daemon restarts.
package statestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry holds the persisted state for a single monitored path.
type Entry struct {
	Path      string      `json:"path"`
	Size      int64       `json:"size"`
	Mode      os.FileMode `json:"mode"`
	ModTime   time.Time   `json:"mod_time"`
	Checksum  string      `json:"checksum,omitempty"`
	RecordedAt time.Time  `json:"recorded_at"`
}

// Store persists and retrieves state entries from a JSON file on disk.
type Store struct {
	mu      sync.RWMutex
	path    string
	entries map[string]Entry
}

// New creates a Store backed by the given file path.
// The directory is created if it does not exist.
// If the file already exists its contents are loaded.
func New(filePath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return nil, fmt.Errorf("statestore: create dir: %w", err)
	}
	s := &Store{path: filePath, entries: make(map[string]Entry)}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("statestore: load: %w", err)
	}
	return s, nil
}

// Set stores an entry for the given path, overwriting any previous value.
func (s *Store) Set(e Entry) error {
	if e.Path == "" {
		return fmt.Errorf("statestore: path must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[e.Path] = e
	return s.flush()
}

// Get retrieves the entry for path. Returns false if not found.
func (s *Store) Get(path string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[path]
	return e, ok
}

// Delete removes the entry for path. It is a no-op if the path is absent.
func (s *Store) Delete(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, path)
	return s.flush()
}

// All returns a copy of all stored entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}

func (s *Store) load() error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&s.entries)
}

func (s *Store) flush() error {
	tmp := s.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(s.entries); err != nil {
		f.Close()
		return err
	}
	f.Close()
	return os.Rename(tmp, s.path)
}

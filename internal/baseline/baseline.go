// Package baseline manages the reference snapshots used to detect drift.
// It persists the last-known-good state for each monitored path so that
// comparisons survive daemon restarts.
package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry holds the reference snapshot for a single file path.
type Entry struct {
	Path    string      `json:"path"`
	Size    int64       `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime time.Time   `json:"mod_time"`
	Hash    string      `json:"hash,omitempty"`
	SetAt   time.Time   `json:"set_at"`
}

// Store persists baseline entries to disk and provides thread-safe access.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
	file    string
}

// New creates a Store backed by the given file path.
// If the file already exists its contents are loaded.
func New(filePath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return nil, fmt.Errorf("baseline: mkdir: %w", err)
	}
	s := &Store{entries: make(map[string]Entry), file: filePath}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// Set records entry as the current baseline for entry.Path and persists the store.
func (s *Store) Set(entry Entry) error {
	entry.SetAt = time.Now()
	s.mu.Lock()
	s.entries[entry.Path] = entry
	s.mu.Unlock()
	return s.save()
}

// Get returns the baseline entry for path and whether it exists.
func (s *Store) Get(path string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[path]
	return e, ok
}

// Delete removes the baseline for path and persists the store.
func (s *Store) Delete(path string) error {
	s.mu.Lock()
	delete(s.entries, path)
	s.mu.Unlock()
	return s.save()
}

// All returns a snapshot of all current entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}

func (s *Store) save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.entries, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("baseline: marshal: %w", err)
	}
	return os.WriteFile(s.file, data, 0o644)
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Unmarshal(data, &s.entries)
}

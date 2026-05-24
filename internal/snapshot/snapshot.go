// Package snapshot provides a thread-safe store for the most recent
// file-state snapshots observed by the watcher pipeline.
package snapshot

import (
	"errors"
	"sync"
	"time"
)

// Entry holds the last-known state for a single monitored path.
type Entry struct {
	Path    string
	Size    int64
	Mode    uint32
	ModTime time.Time
	Missing bool
	Updated time.Time
}

// Store is a concurrent map of path → Entry.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an empty Store.
func New() *Store {
	return &Store{entries: make(map[string]Entry)}
}

// Set inserts or replaces the Entry for path.
func (s *Store) Set(e Entry) error {
	if e.Path == "" {
		return errors.New("snapshot: path must not be empty")
	}
	e.Updated = time.Now()
	s.mu.Lock()
	s.entries[e.Path] = e
	s.mu.Unlock()
	return nil
}

// Get returns the Entry for path and whether it was found.
func (s *Store) Get(path string) (Entry, bool) {
	s.mu.RLock()
	e, ok := s.entries[path]
	s.mu.RUnlock()
	return e, ok
}

// Delete removes the entry for path, if present.
func (s *Store) Delete(path string) {
	s.mu.Lock()
	delete(s.entries, path)
	s.mu.Unlock()
}

// All returns a shallow copy of all stored entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	s.mu.RUnlock()
	return out
}

// Len returns the number of tracked paths.
func (s *Store) Len() int {
	s.mu.RLock()
	n := len(s.entries)
	s.mu.RUnlock()
	return n
}

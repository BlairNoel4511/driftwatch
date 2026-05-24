// Package registry maintains a live index of monitored paths and their
// associated metadata, allowing other subsystems to enumerate targets
// without re-reading the configuration on every tick.
package registry

import (
	"fmt"
	"sync"
)

// Entry holds the metadata driftwatch tracks for a single monitored path.
type Entry struct {
	Path    string
	Enabled bool
	Tags    []string
}

// Registry is a thread-safe store of monitored path entries.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{
		entries: make(map[string]Entry),
	}
}

// Register adds or replaces the entry for the given path.
// It returns an error if path is empty.
func (r *Registry) Register(e Entry) error {
	if e.Path == "" {
		return fmt.Errorf("registry: path must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[e.Path] = e
	return nil
}

// Deregister removes the entry for path. It is a no-op if the path is
// not present.
func (r *Registry) Deregister(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, path)
}

// Get returns the Entry for path and whether it was found.
func (r *Registry) Get(path string) (Entry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[path]
	return e, ok
}

// All returns a snapshot of every registered entry.
func (r *Registry) All() []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Entry, 0, len(r.entries))
	for _, e := range r.entries {
		out = append(out, e)
	}
	return out
}

// Len returns the number of registered paths.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}

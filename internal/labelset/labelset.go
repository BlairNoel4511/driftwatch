// Package labelset provides a thread-safe key/value label store for
// attaching arbitrary metadata to monitored paths.
package labelset

import (
	"fmt"
	"sync"
)

// LabelSet holds string key/value labels keyed by monitored path.
type LabelSet struct {
	mu     sync.RWMutex
	store  map[string]map[string]string
}

// New returns an initialised LabelSet.
func New() *LabelSet {
	return &LabelSet{
		store: make(map[string]map[string]string),
	}
}

// Set replaces all labels for path with the provided map.
// A nil map clears all labels for the path.
func (ls *LabelSet) Set(path string, labels map[string]string) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	if labels == nil {
		delete(ls.store, path)
		return
	}
	copy := make(map[string]string, len(labels))
	for k, v := range labels {
		copy[k] = v
	}
	ls.store[path] = copy
}

// Put sets a single label key/value for path, merging with existing labels.
func (ls *LabelSet) Put(path, key, value string) error {
	if key == "" {
		return fmt.Errorf("labelset: key must not be empty")
	}
	ls.mu.Lock()
	defer ls.mu.Unlock()
	if ls.store[path] == nil {
		ls.store[path] = make(map[string]string)
	}
	ls.store[path][key] = value
	return nil
}

// Get returns a snapshot of all labels for path and whether any exist.
func (ls *LabelSet) Get(path string) (map[string]string, bool) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	src, ok := ls.store[path]
	if !ok {
		return nil, false
	}
	copy := make(map[string]string, len(src))
	for k, v := range src {
		copy[k] = v
	}
	return copy, true
}

// Delete removes a single label key from path.
// It is a no-op if path or key does not exist.
func (ls *LabelSet) Delete(path, key string) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	if m, ok := ls.store[path]; ok {
		delete(m, key)
		if len(m) == 0 {
			delete(ls.store, path)
		}
	}
}

// Paths returns the list of paths that have at least one label.
func (ls *LabelSet) Paths() []string {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	paths := make([]string, 0, len(ls.store))
	for p := range ls.store {
		paths = append(paths, p)
	}
	return paths
}

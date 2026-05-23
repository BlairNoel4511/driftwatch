// Package tagger provides label-based tagging for monitored file paths,
// allowing drift events to be annotated with user-defined metadata.
package tagger

import "sync"

// Tagger maps file paths to a set of string tags.
type Tagger struct {
	mu   sync.RWMutex
	tags map[string][]string
}

// New returns an initialised Tagger.
func New() *Tagger {
	return &Tagger{
		tags: make(map[string][]string),
	}
}

// Set replaces all tags for the given path.
func (t *Tagger) Set(path string, tags []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	copy := make([]string, len(tags))
	for i, v := range tags {
		copy[i] = v
	}
	t.tags[path] = copy
}

// Add appends tags to the existing set for path, deduplicating as it goes.
func (t *Tagger) Add(path string, tags ...string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	existing := t.tags[path]
	set := make(map[string]struct{}, len(existing))
	for _, v := range existing {
		set[v] = struct{}{}
	}
	for _, v := range tags {
		if _, ok := set[v]; !ok {
			existing = append(existing, v)
			set[v] = struct{}{}
		}
	}
	t.tags[path] = existing
}

// Get returns the tags associated with path. Returns nil if none are set.
func (t *Tagger) Get(path string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	v := t.tags[path]
	if v == nil {
		return nil
	}
	copy := make([]string, len(v))
	for i, s := range v {
		copy[i] = s
	}
	return copy
}

// Remove deletes all tags for path.
func (t *Tagger) Remove(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.tags, path)
}

// Paths returns all paths that currently have at least one tag.
func (t *Tagger) Paths() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]string, 0, len(t.tags))
	for k := range t.tags {
		out = append(out, k)
	}
	return out
}

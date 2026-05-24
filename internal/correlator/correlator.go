// Package correlator groups related drift events by a shared correlation key,
// enabling downstream consumers to reason about clusters of related changes
// rather than individual file-level events.
package correlator

import (
	"sync"
	"time"

	"github.com/driftwatch/driftwatch/internal/watcher"
)

// Group holds a set of drift events that share the same correlation key.
type Group struct {
	Key    string
	Events []watcher.Event
	First  time.Time
	Last   time.Time
}

// KeyFunc derives a correlation key from a drift event. Callers supply their
// own implementation — e.g. grouping by directory, owner tag, or service label.
type KeyFunc func(e watcher.Event) string

// Correlator accumulates drift events and exposes them grouped by key.
type Correlator struct {
	mu      sync.Mutex
	keyFunc KeyFunc
	groups  map[string]*Group
}

// New creates a Correlator that uses keyFunc to assign events to groups.
// Panics if keyFunc is nil.
func New(keyFunc KeyFunc) *Correlator {
	if keyFunc == nil {
		panic("correlator: keyFunc must not be nil")
	}
	return &Correlator{
		keyFunc: keyFunc,
		groups:  make(map[string]*Group),
	}
}

// Record adds e to the appropriate group, creating the group if necessary.
func (c *Correlator) Record(e watcher.Event) {
	key := c.keyFunc(e)
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	g, ok := c.groups[key]
	if !ok {
		g = &Group{Key: key, First: now}
		c.groups[key] = g
	}
	g.Events = append(g.Events, e)
	g.Last = now
}

// Groups returns a snapshot of all current groups. The returned slice is a
// copy; modifications do not affect the Correlator's internal state.
func (c *Correlator) Groups() []Group {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]Group, 0, len(c.groups))
	for _, g := range c.groups {
		copy := *g
		copy.Events = append([]watcher.Event(nil), g.Events...)
		out = append(out, copy)
	}
	return out
}

// Clear removes all accumulated groups.
func (c *Correlator) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.groups = make(map[string]*Group)
}

// Len returns the number of distinct correlation groups currently held.
func (c *Correlator) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.groups)
}

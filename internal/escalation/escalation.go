// Package escalation provides a mechanism to escalate drift alerts
// when a path has been in a drifted state for longer than a configured threshold.
package escalation

import (
	"sync"
	"time"
)

// Level represents the severity level of an escalation.
type Level int

const (
	LevelNone     Level = iota
	LevelWarning        // drift persisted beyond warning threshold
	LevelCritical       // drift persisted beyond critical threshold
)

// String returns a human-readable label for the level.
func (l Level) String() string {
	switch l {
	case LevelWarning:
		return "warning"
	case LevelCritical:
		return "critical"
	default:
		return "none"
	}
}

// entry tracks when drift was first observed for a path.
type entry struct {
	firstSeen time.Time
}

// Escalator tracks how long each path has been drifted and returns
// the appropriate escalation level.
type Escalator struct {
	mu       sync.Mutex
	warning  time.Duration
	critical time.Duration
	paths    map[string]entry
	now      func() time.Time
}

// New creates an Escalator. warning and critical define how long a path
// must remain drifted before each level is triggered.
// Panics if warning or critical are zero, or warning >= critical.
func New(warning, critical time.Duration) *Escalator {
	if warning <= 0 {
		panic("escalation: warning threshold must be positive")
	}
	if critical <= 0 {
		panic("escalation: critical threshold must be positive")
	}
	if warning >= critical {
		panic("escalation: warning threshold must be less than critical")
	}
	return &Escalator{
		warning:  warning,
		critical: critical,
		paths:    make(map[string]entry),
		now:      time.Now,
	}
}

// Observe records that path is currently drifted and returns its escalation level.
func (e *Escalator) Observe(path string) Level {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := e.now()
	if _, ok := e.paths[path]; !ok {
		e.paths[path] = entry{firstSeen: now}
	}

	duration := now.Sub(e.paths[path].firstSeen)
	switch {
	case duration >= e.critical:
		return LevelCritical
	case duration >= e.warning:
		return LevelWarning
	default:
		return LevelNone
	}
}

// Resolve clears the drift record for path (drift has been resolved).
func (e *Escalator) Resolve(path string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.paths, path)
}

// Active returns all paths currently being tracked.
func (e *Escalator) Active() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	paths := make([]string, 0, len(e.paths))
	for p := range e.paths {
		paths = append(paths, p)
	}
	return paths
}

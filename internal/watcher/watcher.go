package watcher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// FileState holds the last known state of a watched file.
type FileState struct {
	Path    string
	Checksum string
	ModTime  time.Time
}

// DriftEvent is emitted when a file's state diverges from its snapshot.
type DriftEvent struct {
	Path        string
	Expected    string
	Actual      string
	DetectedAt  time.Time
}

// Watcher monitors a set of file paths for changes.
type Watcher struct {
	mu       sync.RWMutex
	snapshot map[string]FileState
	Events   chan DriftEvent
	Errors   chan error
}

// New creates a new Watcher.
func New() *Watcher {
	return &Watcher{
		snapshot: make(map[string]FileState),
		Events:   make(chan DriftEvent, 16),
		Errors:   make(chan error, 8),
	}
}

// Snapshot records the current state of the given paths as the baseline.
func (w *Watcher) Snapshot(paths []string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, p := range paths {
		state, err := statFile(p)
		if err != nil {
			return fmt.Errorf("snapshot %s: %w", p, err)
		}
		w.snapshot[p] = state
	}
	return nil
}

// Check compares current file states against the snapshot and emits DriftEvents.
func (w *Watcher) Check() {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for path, expected := range w.snapshot {
		actual, err := statFile(path)
		if err != nil {
			w.Errors <- fmt.Errorf("check %s: %w", path, err)
			continue
		}
		if actual.Checksum != expected.Checksum {
			w.Events <- DriftEvent{
				Path:       path,
				Expected:   expected.Checksum,
				Actual:     actual.Checksum,
				DetectedAt: time.Now(),
			}
		}
	}
}

// statFile returns a FileState for the given path.
func statFile(path string) (FileState, error) {
	f, err := os.Open(path)
	if err != nil {
		return FileState{}, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return FileState{}, err
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return FileState{}, err
	}

	return FileState{
		Path:     path,
		Checksum: fmt.Sprintf("%x", h.Sum(nil)),
		ModTime:  info.ModTime(),
	}, nil
}

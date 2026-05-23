// Package audit provides a structured audit log for drift events,
// recording when drift was detected, acknowledged, or resolved.
package audit

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// EventKind classifies the type of audit entry.
type EventKind string

const (
	KindDetected    EventKind = "detected"
	KindAcknowledged EventKind = "acknowledged"
	KindResolved    EventKind = "resolved"
)

// Entry is a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Kind      EventKind `json:"kind"`
	Path      string    `json:"path"`
	Detail    string    `json:"detail,omitempty"`
}

// Log writes structured audit entries to an io.Writer.
type Log struct {
	mu  sync.Mutex
	out io.Writer
	enc *json.Encoder
}

// New returns a Log that writes JSON lines to w.
// If w is nil, os.Stderr is used.
func New(w io.Writer) *Log {
	if w == nil {
		w = os.Stderr
	}
	return &Log{
		out: w,
		enc: json.NewEncoder(w),
	}
}

// Record appends an Entry to the log.
func (l *Log) Record(kind EventKind, path, detail string) error {
	e := Entry{
		Timestamp: time.Now().UTC(),
		Kind:      kind,
		Path:      path,
		Detail:    detail,
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enc.Encode(e)
}

// Detected is a convenience wrapper for KindDetected entries.
func (l *Log) Detected(path, detail string) error {
	return l.Record(KindDetected, path, detail)
}

// Acknowledged is a convenience wrapper for KindAcknowledged entries.
func (l *Log) Acknowledged(path, detail string) error {
	return l.Record(KindAcknowledged, path, detail)
}

// Resolved is a convenience wrapper for KindResolved entries.
func (l *Log) Resolved(path, detail string) error {
	return l.Record(KindResolved, path, detail)
}

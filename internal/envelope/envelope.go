// Package envelope wraps a drift event with routing metadata such as
// severity, destination topic, and trace identifiers so that downstream
// sinks can make routing decisions without inspecting the raw payload.
package envelope

import (
	"fmt"
	"time"

	"github.com/driftwatch/driftwatch/internal/watcher"
)

// Severity classifies how urgent a drift event is.
type Severity int

const (
	SeverityInfo     Severity = iota // informational, no action required
	SeverityWarning                  // degraded, should be investigated
	SeverityCritical                 // critical, immediate action required
)

// String returns a human-readable label for the severity level.
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityCritical:
		return "critical"
	default:
		return fmt.Sprintf("severity(%d)", int(s))
	}
}

// Envelope carries a drift event together with routing metadata.
type Envelope struct {
	// ID is a unique identifier for this envelope instance.
	ID string
	// Topic is the logical destination channel (e.g. "drift.critical").
	Topic string
	// Severity indicates the urgency level of the enclosed event.
	Severity Severity
	// Event is the underlying drift event being wrapped.
	Event watcher.Event
	// CreatedAt records when the envelope was constructed.
	CreatedAt time.Time
	// Attempt tracks how many delivery attempts have been made.
	Attempt int
}

// New constructs an Envelope for the given event, assigning the provided
// id, topic, and severity. CreatedAt is set to the current UTC time.
func New(id, topic string, severity Severity, event watcher.Event) Envelope {
	if id == "" {
		panic("envelope: id must not be empty")
	}
	if topic == "" {
		panic("envelope: topic must not be empty")
	}
	return Envelope{
		ID:        id,
		Topic:     topic,
		Severity:  severity,
		Event:     event,
		CreatedAt: time.Now().UTC(),
		Attempt:   0,
	}
}

// Retry returns a copy of the envelope with Attempt incremented by one.
func (e Envelope) Retry() Envelope {
	e.Attempt++
	return e
}

package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/driftwatch/internal/watcher"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Event represents a formatted drift alert.
type Event struct {
	Timestamp time.Time
	Level     Level
	Path      string
	Message   string
}

// Notifier sends drift alerts to a destination.
type Notifier struct {
	out io.Writer
}

// New creates a Notifier that writes to the given writer.
// If w is nil, os.Stderr is used.
func New(w io.Writer) *Notifier {
	if w == nil {
		w = os.Stderr
	}
	return &Notifier{out: w}
}

// Notify formats a DriftEvent and writes it to the configured writer.
func (n *Notifier) Notify(de watcher.DriftEvent) error {
	ev := Event{
		Timestamp: time.Now().UTC(),
		Level:     LevelWarn,
		Path:      de.Path,
		Message:   buildMessage(de),
	}
	_, err := fmt.Fprintf(n.out, "[%s] %s %s: %s\n",
		ev.Timestamp.Format(time.RFC3339),
		ev.Level,
		ev.Path,
		ev.Message,
	)
	return err
}

func buildMessage(de watcher.DriftEvent) string {
	if de.Current == nil {
		return "file missing — was present at snapshot"
	}
	if de.Snapshot.Size != de.Current.Size {
		return fmt.Sprintf("size changed: %d → %d bytes", de.Snapshot.Size, de.Current.Size)
	}
	if de.Snapshot.ModTime != de.Current.ModTime {
		return fmt.Sprintf("mtime changed: %s → %s",
			de.Snapshot.ModTime.Format(time.RFC3339),
			de.Current.ModTime.Format(time.RFC3339))
	}
	if de.Snapshot.Mode != de.Current.Mode {
		return fmt.Sprintf("mode changed: %s → %s", de.Snapshot.Mode, de.Current.Mode)
	}
	return "unknown drift detected"
}

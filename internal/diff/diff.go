// Package diff computes the human-readable delta between two file snapshots.
package diff

import (
	"fmt"
	"strings"

	"github.com/user/driftwatch/internal/watcher"
)

// Field names for changed attributes.
const (
	FieldSize    = "size"
	FieldMode    = "mode"
	FieldModTime = "mod_time"
	FieldMissing = "missing"
)

// Change describes a single attribute that diverged between two snapshots.
type Change struct {
	Field    string
	OldValue string
	NewValue string
}

// String returns a concise human-readable representation of the change.
func (c Change) String() string {
	return fmt.Sprintf("%s: %s → %s", c.Field, c.OldValue, c.NewValue)
}

// Result holds all detected changes for a single file path.
type Result struct {
	Path    string
	Changes []Change
}

// HasDrift reports whether any changes were detected.
func (r Result) HasDrift() bool {
	return len(r.Changes) > 0
}

// Summary returns a single-line summary of all changes.
func (r Result) Summary() string {
	parts := make([]string, len(r.Changes))
	for i, c := range r.Changes {
		parts[i] = c.String()
	}
	return fmt.Sprintf("%s: [%s]", r.Path, strings.Join(parts, ", "))
}

// Compare returns a Result describing how current diverges from baseline.
// If baseline is nil the file is considered newly missing or appeared.
func Compare(path string, baseline, current *watcher.Snapshot) Result {
	res := Result{Path: path}

	if baseline == nil && current == nil {
		return res
	}

	if current == nil {
		res.Changes = append(res.Changes, Change{
			Field:    FieldMissing,
			OldValue: "present",
			NewValue: "missing",
		})
		return res
	}

	if baseline == nil {
		res.Changes = append(res.Changes, Change{
			Field:    FieldMissing,
			OldValue: "missing",
			NewValue: "present",
		})
		return res
	}

	if baseline.Size != current.Size {
		res.Changes = append(res.Changes, Change{
			Field:    FieldSize,
			OldValue: fmt.Sprintf("%d", baseline.Size),
			NewValue: fmt.Sprintf("%d", current.Size),
		})
	}

	if baseline.Mode != current.Mode {
		res.Changes = append(res.Changes, Change{
			Field:    FieldMode,
			OldValue: baseline.Mode.String(),
			NewValue: current.Mode.String(),
		})
	}

	if !baseline.ModTime.Equal(current.ModTime) {
		res.Changes = append(res.Changes, Change{
			Field:    FieldModTime,
			OldValue: baseline.ModTime.UTC().Format("2006-01-02T15:04:05Z"),
			NewValue: current.ModTime.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	return res
}

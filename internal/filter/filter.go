// Package filter provides path-based filtering for drift targets.
// It supports glob patterns to include or exclude files from monitoring.
package filter

import (
	"path/filepath"
	"strings"
)

// Filter holds compiled include and exclude glob patterns.
type Filter struct {
	includes []string
	excludes []string
}

// New creates a Filter from the given include and exclude pattern slices.
// If includes is empty, all paths are considered included by default.
func New(includes, excludes []string) *Filter {
	return &Filter{
		includes: includes,
		excludes: excludes,
	}
}

// Allow reports whether the given path should be monitored.
// A path is allowed when it matches at least one include pattern (or no
// include patterns are defined) and does not match any exclude pattern.
func (f *Filter) Allow(path string) (bool, error) {
	if len(f.excludes) > 0 {
		for _, pattern := range f.excludes {
			matched, err := matchPattern(pattern, path)
			if err != nil {
				return false, err
			}
			if matched {
				return false, nil
			}
		}
	}

	if len(f.includes) == 0 {
		return true, nil
	}

	for _, pattern := range f.includes {
		matched, err := matchPattern(pattern, path)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}

// matchPattern checks a single glob pattern against a path.
// It tries both a direct match and a base-name match for convenience.
func matchPattern(pattern, path string) (bool, error) {
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return false, err
	}
	if matched {
		return true, nil
	}
	// Also match against the base name so patterns like "*.yaml" work
	// without requiring a full path prefix.
	if !strings.ContainsRune(pattern, '/') {
		return filepath.Match(pattern, filepath.Base(path))
	}
	return false, nil
}

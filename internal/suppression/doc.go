// Package suppression provides duplicate-alert suppression for driftwatch.
//
// A Suppressor tracks the last time an alert was emitted for each monitored
// path. Subsequent alerts for the same path are suppressed until the
// configured window has elapsed, preventing alert storms when a file
// remains in a drifted state across multiple polling cycles.
//
// Example usage:
//
//	sup := suppression.New(5 * time.Minute)
//	if sup.Allow(path) {
//		// emit alert
//	}
package suppression

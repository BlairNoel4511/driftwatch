// Package debounce provides a path-keyed debouncer for drift events.
//
// When infrastructure files change rapidly (e.g. during a rolling deploy),
// the watcher may emit multiple drift events for the same path in quick
// succession. The Debouncer coalesces these into a single callback invocation
// that fires only after a configurable quiet period has elapsed, reducing
// alert noise and downstream load on notifiers and reporters.
//
// Usage:
//
//	d := debounce.New(500*time.Millisecond, func(path string) {
//		// handle drift for path
//	})
//	d.Trigger("/etc/app/config.yaml")
package debounce

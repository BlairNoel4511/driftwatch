// Package backoff implements a per-key exponential back-off tracker.
//
// Typical usage:
//
//	b := backoff.New(500*time.Millisecond, 30*time.Second)
//
//	// on failure:
//	wait := b.Next("/etc/myapp/config.yaml")
//	time.Sleep(wait)
//
//	// on success:
//	b.Reset("/etc/myapp/config.yaml")
package backoff

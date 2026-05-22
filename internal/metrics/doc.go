// Package metrics provides lightweight atomic counters for tracking
// driftwatch runtime activity, including the number of file checks
// performed, drift events detected, alerts dispatched, and errors
// encountered during operation.
//
// Counters are safe for concurrent use. A Snapshot method provides a
// point-in-time read of all values suitable for logging or export.
package metrics

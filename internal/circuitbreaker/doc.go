// Package circuitbreaker provides a per-path circuit breaker for driftwatch.
//
// When a monitored path produces consecutive check failures (e.g. the file
// becomes unreadable or the watcher returns errors), the circuit opens and
// further processing for that path is skipped until a cooldown period elapses.
// This prevents alert storms and excessive retries on persistently broken
// targets while still recovering automatically once the underlying issue
// is resolved.
package circuitbreaker

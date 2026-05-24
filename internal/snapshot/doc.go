// Package snapshot provides a concurrency-safe in-memory store for the
// most recent observed state of each monitored file.
//
// The watcher pipeline writes an Entry whenever it polls a target path;
// downstream components (diff, reporter, healthcheck) read from the store
// to compare declared vs live state without re-stating the filesystem.
package snapshot

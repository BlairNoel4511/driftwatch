// Package baseline manages reference snapshots (the "last-known-good" state)
// for every path monitored by driftwatch.
//
// A baseline.Store is backed by a single JSON file so that the reference
// state survives daemon restarts.  Callers set an entry after a successful
// check and compare live state against it to determine whether drift has
// occurred.
package baseline

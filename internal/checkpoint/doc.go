// Package checkpoint provides a lightweight persistence layer for driftwatch.
//
// It serialises file-state snapshots to JSON on disk so that the daemon can
// distinguish genuine drift from the expected state captured before a restart.
//
// Usage:
//
//	store, err := checkpoint.New("/var/lib/driftwatch/checkpoints")
//	if err != nil { ... }
//
//	// Persist a snapshot after a successful baseline read.
//	store.Save(checkpoint.Snapshot{Path: "/etc/hosts", Size: 312, ...})
//
//	// Retrieve it on the next run.
//	snap, ok, err := store.Load("/etc/hosts")
package checkpoint

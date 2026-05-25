// Package statestore provides a lightweight, file-backed persistent store
// for the last-known state of each monitored path.
//
// Entries are serialised as JSON so they survive daemon restarts, allowing
// driftwatch to distinguish a genuinely changed file from one that was
// simply unobserved while the process was down.
//
// Typical usage:
//
//	st, err := statestore.New("/var/lib/driftwatch/state.json")
//	if err != nil { ... }
//
//	entry, ok := st.Get("/etc/hosts")
//	if !ok { /* first time we have seen this path */ }
package statestore

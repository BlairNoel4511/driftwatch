// Package fingerprint derives a stable, opaque identity token for each
// monitored file by hashing its observable metadata (path, size, mode,
// modification time) together with its content checksum.
//
// # Overview
//
// A [Fingerprint] is a fixed-size digest that captures the combined state of a
// file at a point in time. Two fingerprints are equal if and only if every
// observed attribute – path, size, permission bits, modification time, and
// content hash – matches exactly.
//
// # Usage
//
// Create a [Fingerprinter] once and reuse it across calls:
//
//	 fp := fingerprint.New()
//	 prev, err := fp.Compute("/etc/config.yaml")
//	 // ... later ...
//	 changed, err := fp.Changed("/etc/config.yaml", prev)
//	 if changed {
//	 	// file state has shifted since last observation
//	 }
//
// Use [Fingerprinter.Changed] to detect whether a file's combined state has
// shifted since it was last observed, without needing to diff individual
// fields.
package fingerprint

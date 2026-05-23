// Package fingerprint derives a stable, opaque identity token for each
// monitored file by hashing its observable metadata (path, size, mode,
// modification time) together with its content checksum.
//
// Use Fingerprinter.Changed to detect whether a file's combined state has
// shifted since it was last observed, without needing to diff individual
// fields.
package fingerprint

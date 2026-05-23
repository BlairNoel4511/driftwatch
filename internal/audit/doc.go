// Package audit provides a structured, append-only audit log for
// driftwatch events. Each entry is written as a JSON line and captures
// the event kind (detected, acknowledged, resolved), the affected file
// path, and an optional human-readable detail string.
//
// Usage:
//
//	log := audit.New(os.Stdout)
//	log.Detected("/etc/nginx/nginx.conf", "size changed 1024 -> 2048")
//	log.Resolved("/etc/nginx/nginx.conf", "snapshot refreshed")
package audit

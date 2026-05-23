// Package digest provides SHA-256 content hashing for driftwatch monitored
// files. It supplements mtime/size checks with a content-level comparison so
// that touch-only changes do not trigger false drift alerts, and byte-level
// modifications are never missed even when a file's size stays the same.
package digest

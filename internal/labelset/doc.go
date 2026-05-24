// Package labelset provides a thread-safe, path-keyed label store used to
// attach arbitrary string metadata (e.g. environment, owner, tier) to
// monitored infrastructure config files.
//
// Labels are stored in memory and are not persisted across restarts. They are
// intended to enrich drift events and alert messages with human-readable
// context that is not present in the file itself.
//
// Example:
//
//	ls := labelset.New()
//	ls.Put("/etc/nginx/nginx.conf", "owner", "platform-team")
//	ls.Put("/etc/nginx/nginx.conf", "env",   "production")
//	labels, _ := ls.Get("/etc/nginx/nginx.conf")
package labelset

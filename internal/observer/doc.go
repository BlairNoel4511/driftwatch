// Package observer tracks the age of active drift events and derives
// a severity level (none, warning, critical) based on how long each
// monitored path has remained in a drifted state.
//
// Typical usage:
//
//	obs := observer.New(5*time.Minute, 30*time.Minute)
//
//	// on every drift event:
//	sev := obs.Observe("/etc/nginx/nginx.conf")
//
//	// when drift is resolved:
//	obs.Resolve("/etc/nginx/nginx.conf")
package observer

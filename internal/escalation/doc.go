// Package escalation tracks how long individual paths have remained in a
// drifted state and promotes alerts to higher severity levels (warning,
// critical) once configurable age thresholds are exceeded.
//
// Typical usage:
//
//	e := escalation.New(5*time.Minute, 30*time.Minute)
//	level := e.Observe("/etc/hosts") // LevelNone, LevelWarning, or LevelCritical
//	e.Resolve("/etc/hosts")          // clear when drift is remediated
package escalation

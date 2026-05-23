// Package silencer provides temporary alert suppression for specific
// monitored paths. It is useful during planned maintenance windows or
// when an operator has acknowledged a drift condition and wants to
// prevent repeated alerts while a fix is being applied.
//
// Silences are time-bounded and expire automatically. Calling Silence
// again on an already-silenced path extends the window to the new
// expiry, whichever is later.
package silencer

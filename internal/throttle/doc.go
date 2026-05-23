// Package throttle implements a fixed-window token-bucket throttle for
// controlling the rate of drift alerts on a per-path basis.
//
// Unlike suppression (which silences repeated identical events) and ratelimit
// (which enforces a cooldown between any two events), throttle allows a
// configurable burst of events within each time window before blocking further
// events until the window resets.
//
// Typical usage:
//
//	th := throttle.New(time.Minute, 5)
//	if th.Allow(path) {
//		// emit alert
//	}
package throttle

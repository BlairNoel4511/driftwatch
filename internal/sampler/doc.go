// Package sampler provides probabilistic sampling for drift events.
//
// A Sampler is constructed with a default pass-through rate and optional
// per-path overrides. Callers invoke Allow(path) before forwarding an event;
// the sampler draws a uniform random variate and compares it against the
// configured rate, returning true when the event should proceed.
//
// Example:
//
//	src := rand.NewSource(time.Now().UnixNano())
//	s := sampler.New(0.1, src)          // forward ~10 % of all events
//	_ = s.SetRate("/etc/hosts", 1.0)    // always forward /etc/hosts changes
//	if s.Allow(path) {
//	    notifier.Dispatch(event)
//	}
package sampler

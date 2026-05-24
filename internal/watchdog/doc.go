// Package watchdog provides a liveness watchdog for driftwatch check loops.
//
// A Watchdog expects periodic Kick calls (one per check cycle). If the
// deadline elapses without a kick — indicating a stalled or crashed loop —
// the registered Handler is invoked with the amount of time overdue.
//
// Typical usage:
//
//	wd := watchdog.New(30*time.Second, func(missed time.Duration) {
//		log.Printf("check loop stalled by %s", missed)
//	})
//	go wd.Run(ctx)
//	// inside check loop:
//	wd.Kick()
package watchdog

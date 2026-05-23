// Package rollup provides a time-window batching layer for drift events.
//
// Instead of forwarding every DriftEvent individually, Rollup accumulates
// events over a configurable window and delivers them as a single Batch to
// a Handler. This reduces downstream alert noise when many files change at
// the same time (e.g. during a deployment).
//
// Basic usage:
//
//	r := rollup.New(5*time.Second, func(b rollup.Batch) {
//		fmt.Printf("flushed %d events\n", len(b.Events))
//	})
//	r.Run(ctx, driftCh)
package rollup

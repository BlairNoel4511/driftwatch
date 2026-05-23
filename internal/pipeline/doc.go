// Package pipeline wires the watcher, diff, filter, alerter, notifier,
// checkpoint, and metrics components into a single Run method that
// represents one complete drift-detection cycle.
//
// Typical usage:
//
//	p, err := pipeline.New(pipeline.Config{
//		Watcher:    w,
//		Filter:     f,
//		Alerter:    a,
//		Notifier:   n,
//		Checkpoint: cp,
//		Metrics:    m,
//	})
//	count, err := p.Run(ctx)
package pipeline

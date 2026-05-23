// Package notifier fans out drift events to registered sinks and records
// them in a history ring-buffer.
package notifier

import (
	"context"
	"log"

	"github.com/driftwatch/driftwatch/internal/alert"
	"github.com/driftwatch/driftwatch/internal/history"
	"github.com/driftwatch/driftwatch/internal/watcher"
)

// Sink is any destination that can receive a drift notification.
type Sink interface {
	Notify(event watcher.DriftEvent) error
}

// Notifier dispatches DriftEvents to one or more Sinks and keeps a History.
type Notifier struct {
	sinks   []Sink
	history *history.History
	events  <-chan watcher.DriftEvent
}

// New creates a Notifier that reads from events and writes to the provided sinks.
func New(events <-chan watcher.DriftEvent, h *history.History, sinks ...Sink) *Notifier {
	return &Notifier{
		sinks:   sinks,
		history: h,
		events:  events,
	}
}

// Run processes events until ctx is cancelled.
func (n *Notifier) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-n.events:
			if !ok {
				return
			}
			n.dispatch(ev)
		}
	}
}

func (n *Notifier) dispatch(ev watcher.DriftEvent) {
	n.history.Record(ev)
	for _, s := range n.sinks {
		if err := s.Notify(ev); err != nil {
			log.Printf("notifier: sink error: %v", err)
		}
	}
}

// AlertSink wraps an *alert.Alerter to satisfy the Sink interface.
type AlertSink struct {
	alerter *alert.Alerter
}

// NewAlertSink creates an AlertSink backed by the given Alerter.
func NewAlertSink(a *alert.Alerter) *AlertSink {
	return &AlertSink{alerter: a}
}

// Notify forwards the event to the underlying Alerter.
func (s *AlertSink) Notify(ev watcher.DriftEvent) error {
	return s.alerter.Notify(ev)
}

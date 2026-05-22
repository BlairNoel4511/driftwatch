// Package notifier dispatches drift events to one or more alert sinks.
package notifier

import (
	"context"
	"log"

	"github.com/yourorg/driftwatch/internal/alert"
	"github.com/yourorg/driftwatch/internal/history"
	"github.com/yourorg/driftwatch/internal/watcher"
)

// Sink is anything that can receive a drift notification.
type Sink interface {
	Notify(event watcher.DriftEvent) error
}

// Notifier fans out drift events to registered sinks and records them in history.
type Notifier struct {
	sinks   []Sink
	history *history.History
	logger  *log.Logger
}

// New creates a Notifier. If logger is nil, output goes to stderr via alert defaults.
func New(h *history.History, logger *log.Logger, sinks ...Sink) *Notifier {
	if logger == nil {
		logger = log.Default()
	}
	return &Notifier{
		sinks:   sinks,
		history: h,
		logger:  logger,
	}
}

// AddSink appends a sink at runtime.
func (n *Notifier) AddSink(s Sink) {
	n.sinks = append(n.sinks, s)
}

// Run reads drift events from ch until ctx is cancelled.
func (n *Notifier) Run(ctx context.Context, ch <-chan watcher.DriftEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			n.dispatch(event)
		}
	}
}

func (n *Notifier) dispatch(event watcher.DriftEvent) {
	if n.history != nil {
		n.history.Record(event)
	}
	for _, s := range n.sinks {
		if err := s.Notify(event); err != nil {
			n.logger.Printf("notifier: sink error for %s: %v", event.Path, err)
		}
	}
}

// AlertSinkAdapter wraps *alert.Alert so it satisfies Sink.
type AlertSinkAdapter struct {
	a *alert.Alert
}

// NewAlertSink creates an AlertSinkAdapter from an existing alert.Alert.
func NewAlertSink(a *alert.Alert) *AlertSinkAdapter {
	return &AlertSinkAdapter{a: a}
}

// Notify implements Sink.
func (as *AlertSinkAdapter) Notify(event watcher.DriftEvent) error {
	return as.a.Notify(event)
}

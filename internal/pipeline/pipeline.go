// Package pipeline wires together the core drift-detection components into
// a single, reusable processing pipeline.
package pipeline

import (
	"context"
	"log"

	"github.com/yourorg/driftwatch/internal/alert"
	"github.com/yourorg/driftwatch/internal/checkpoint"
	"github.com/yourorg/driftwatch/internal/diff"
	"github.com/yourorg/driftwatch/internal/filter"
	"github.com/yourorg/driftwatch/internal/metrics"
	"github.com/yourorg/driftwatch/internal/notifier"
	"github.com/yourorg/driftwatch/internal/watcher"
)

// Pipeline orchestrates a single drift-check cycle.
type Pipeline struct {
	watcher    *watcher.Watcher
	filter     *filter.Filter
	alerter    *alert.Alerter
	notifier   *notifier.Notifier
	checkpoint *checkpoint.Store
	metrics    *metrics.Metrics
	logger     *log.Logger
}

// Config holds dependencies required to construct a Pipeline.
type Config struct {
	Watcher    *watcher.Watcher
	Filter     *filter.Filter
	Alerter    *alert.Alerter
	Notifier   *notifier.Notifier
	Checkpoint *checkpoint.Store
	Metrics    *metrics.Metrics
	Logger     *log.Logger
}

// New creates a Pipeline from the supplied Config.
func New(cfg Config) *Pipeline {
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	return &Pipeline{
		watcher:    cfg.Watcher,
		filter:     cfg.Filter,
		alerter:    cfg.Alerter,
		notifier:   cfg.Notifier,
		checkpoint: cfg.Checkpoint,
		metrics:    cfg.Metrics,
		logger:     cfg.Logger,
	}
}

// Run executes one full drift-detection cycle and returns the number of
// drifted paths detected.
func (p *Pipeline) Run(ctx context.Context) (int, error) {
	snap, err := p.watcher.Snapshot()
	if err != nil {
		p.metrics.RecordError()
		return 0, err
	}
	p.metrics.RecordCheck()

	prev, _, err := p.checkpoint.Load()
	if err != nil {
		p.metrics.RecordError()
		return 0, err
	}

	events := diff.Compare(prev, snap)
	driftCount := 0

	for _, ev := range events {
		ok, err := p.filter.Allow(ev.Path)
		if err != nil || !ok {
			continue
		}
		p.metrics.RecordDrift()
		driftCount++
		msg := p.alerter.BuildMessage(ev)
		if nErr := p.notifier.Dispatch(ctx, msg); nErr != nil {
			p.logger.Printf("pipeline: dispatch error for %s: %v", ev.Path, nErr)
			p.metrics.RecordError()
		}
		p.metrics.RecordAlert()
	}

	if err := p.checkpoint.Save(snap); err != nil {
		p.logger.Printf("pipeline: checkpoint save error: %v", err)
		p.metrics.RecordError()
	}
	return driftCount, nil
}

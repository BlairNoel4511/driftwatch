package watcher

import (
	"context"
	"log/slog"
	"time"
)

// Poller runs periodic drift checks against a Watcher.
type Poller struct {
	w        *Watcher
	interval time.Duration
	logger   *slog.Logger
}

// NewPoller creates a Poller that checks for drift on the given interval.
func NewPoller(w *Watcher, interval time.Duration, logger *slog.Logger) *Poller {
	if logger == nil {
		logger = slog.Default()
	}
	return &Poller{w: w, interval: interval, logger: logger}
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (p *Poller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	p.logger.Info("poller started", "interval", p.interval)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("poller stopped")
			return
		case <-ticker.C:
			p.w.Check()
		case ev := <-p.w.Events:
			p.logger.Warn("drift detected",
				"path", ev.Path,
				"expected_checksum", ev.Expected,
				"actual_checksum", ev.Actual,
				"detected_at", ev.DetectedAt,
			)
		case err := <-p.w.Errors:
			p.logger.Error("watcher error", "err", err)
		}
	}
}

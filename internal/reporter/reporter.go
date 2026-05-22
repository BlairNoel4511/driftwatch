package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/driftwatch/driftwatch/internal/watcher"
)

// Format controls the output format of drift reports.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// DriftReport represents a single drift detection event.
type DriftReport struct {
	Timestamp time.Time        `json:"timestamp"`
	Path      string           `json:"path"`
	Reason    string           `json:"reason"`
	Baseline  watcher.Snapshot `json:"baseline"`
	Current   watcher.Snapshot `json:"current"`
}

// Reporter writes drift reports to an output sink.
type Reporter struct {
	out    io.Writer
	format Format
}

// New creates a Reporter. If out is nil, os.Stdout is used.
func New(out io.Writer, format Format) *Reporter {
	if out == nil {
		out = os.Stdout
	}
	if format == "" {
		format = FormatText
	}
	return &Reporter{out: out, format: format}
}

// Report writes a drift report for the given path and snapshots.
func (r *Reporter) Report(path, reason string, baseline, current watcher.Snapshot) error {
	rpt := DriftReport{
		Timestamp: time.Now().UTC(),
		Path:      path,
		Reason:    reason,
		Baseline:  baseline,
		Current:   current,
	}

	switch r.format {
	case FormatJSON:
		return r.writeJSON(rpt)
	default:
		return r.writeText(rpt)
	}
}

func (r *Reporter) writeText(rpt DriftReport) error {
	_, err := fmt.Fprintf(
		r.out,
		"[%s] DRIFT DETECTED path=%s reason=%s baseline_size=%d current_size=%d baseline_mode=%s current_mode=%s\n",
		rpt.Timestamp.Format(time.RFC3339),
		rpt.Path,
		rpt.Reason,
		rpt.Baseline.Size,
		rpt.Current.Size,
		rpt.Baseline.Mode,
		rpt.Current.Mode,
	)
	return err
}

func (r *Reporter) writeJSON(rpt DriftReport) error {
	enc := json.NewEncoder(r.out)
	return enc.Encode(rpt)
}

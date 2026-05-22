// Package reporter provides formatted output for drift detection events.
//
// It supports two output formats:
//
//	- text: a human-readable single-line log entry (default)
//	- json: a machine-readable JSON object, one per line
//
// Usage:
//
//	r := reporter.New(os.Stdout, reporter.FormatJSON)
//	r.Report(path, reason, baseline, current)
package reporter

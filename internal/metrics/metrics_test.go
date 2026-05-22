package metrics_test

import (
	"testing"

	"github.com/driftwatch/driftwatch/internal/metrics"
)

func TestNew_EmptyCounters(t *testing.T) {
	m := metrics.New(nil)
	snap := m.Snapshot()
	for k, v := range snap {
		if v != 0 {
			t.Errorf("expected %s=0, got %d", k, v)
		}
	}
}

func TestRecordCheck(t *testing.T) {
	m := metrics.New(nil)
	m.RecordCheck()
	m.RecordCheck()
	if got := m.ChecksTotal.Load(); got != 2 {
		t.Errorf("expected 2 checks, got %d", got)
	}
}

func TestRecordDrift(t *testing.T) {
	m := metrics.New(nil)
	m.RecordDrift()
	if got := m.DriftsTotal.Load(); got != 1 {
		t.Errorf("expected 1 drift, got %d", got)
	}
}

func TestRecordAlert(t *testing.T) {
	m := metrics.New(nil)
	m.RecordAlert()
	m.RecordAlert()
	m.RecordAlert()
	if got := m.AlertsTotal.Load(); got != 3 {
		t.Errorf("expected 3 alerts, got %d", got)
	}
}

func TestRecordError(t *testing.T) {
	m := metrics.New(nil)
	m.RecordError()
	if got := m.ErrorsTotal.Load(); got != 1 {
		t.Errorf("expected 1 error, got %d", got)
	}
}

func TestSnapshot_AllCounters(t *testing.T) {
	m := metrics.New(nil)
	m.RecordCheck()
	m.RecordDrift()
	m.RecordAlert()
	m.RecordError()
	snap := m.Snapshot()
	expected := map[string]int64{
		"checks_total": 1,
		"drifts_total": 1,
		"alerts_total": 1,
		"errors_total": 1,
	}
	for k, want := range expected {
		if got := snap[k]; got != want {
			t.Errorf("snap[%s]: want %d, got %d", k, want, got)
		}
	}
}

func TestLabels_Returned(t *testing.T) {
	m := metrics.New(map[string]string{"env": "test", "host": "localhost"})
	labels := m.Labels()
	if labels["env"] != "test" {
		t.Errorf("expected env=test, got %s", labels["env"])
	}
	if labels["host"] != "localhost" {
		t.Errorf("expected host=localhost, got %s", labels["host"])
	}
}

func TestReset_ZeroesCounters(t *testing.T) {
	m := metrics.New(nil)
	m.RecordCheck()
	m.RecordDrift()
	m.Reset()
	snap := m.Snapshot()
	for k, v := range snap {
		if v != 0 {
			t.Errorf("after Reset, expected %s=0, got %d", k, v)
		}
	}
}

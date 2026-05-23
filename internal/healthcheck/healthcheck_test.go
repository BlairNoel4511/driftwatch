package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/driftwatch/internal/healthcheck"
)

func TestServeHTTP_HealthyResponse(t *testing.T) {
	h := healthcheck.New()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var s healthcheck.Status
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !s.Healthy {
		t.Error("expected healthy=true")
	}
	if s.DriftCount != 0 {
		t.Errorf("expected drift_count=0, got %d", s.DriftCount)
	}
}

func TestRecordDrift_IncrementsDriftCount(t *testing.T) {
	h := healthcheck.New()
	h.RecordDrift()
	h.RecordDrift()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	var s healthcheck.Status
	_ = json.NewDecoder(rec.Body).Decode(&s)

	if s.DriftCount != 2 {
		t.Errorf("expected drift_count=2, got %d", s.DriftCount)
	}
}

func TestRecordCheck_SetsLastCheck(t *testing.T) {
	h := healthcheck.New()

	before := time.Now().Add(-time.Second)
	h.RecordCheck()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	var s healthcheck.Status
	_ = json.NewDecoder(rec.Body).Decode(&s)

	if s.LastCheck.IsZero() {
		t.Fatal("expected last_check to be set")
	}
	if s.LastCheck.Before(before) {
		t.Errorf("last_check %v is before expected lower bound %v", s.LastCheck, before)
	}
}

func TestServeHTTP_ContentTypeJSON(t *testing.T) {
	h := healthcheck.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

// Package healthcheck exposes a lightweight HTTP endpoint that reports
// the operational health of the driftwatch daemon.
//
// Mount the Handler on any http.ServeMux to enable a /healthz route:
//
//	h := healthcheck.New()
//	mux.Handle("/healthz", h)
//
// The handler returns HTTP 200 with a JSON body containing:
//   - healthy: always true while the process is running
//   - last_check: timestamp of the most recent file-state poll
//   - drift_count: cumulative number of drift events observed
//   - uptime: human-readable duration since the daemon started
package healthcheck

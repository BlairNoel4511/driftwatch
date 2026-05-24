// Package envelope provides a thin wrapper around drift events that attaches
// routing metadata — severity, topic, trace ID, and delivery attempt count —
// so that alert sinks, routers, and retry mechanisms can make decisions
// without inspecting raw watcher payloads.
package envelope

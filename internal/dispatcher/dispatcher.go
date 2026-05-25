// Package dispatcher routes drift events to registered handlers based on
// path-prefix rules, enabling per-target notification strategies.
package dispatcher

import (
	"fmt"
	"strings"
	"sync"

	"github.com/yourorg/driftwatch/internal/watcher"
)

// Handler is a function that receives a drift event.
type Handler func(event watcher.Event)

// Route pairs a path prefix with a handler.
type Route struct {
	Prefix  string
	Handler Handler
}

// Dispatcher routes events to handlers whose prefix matches the event path.
type Dispatcher struct {
	mu     sync.RWMutex
	routes []Route
	fallback Handler
}

// New returns a Dispatcher with an optional fallback handler invoked when no
// route matches. fallback may be nil.
func New(fallback Handler) *Dispatcher {
	return &Dispatcher{fallback: fallback}
}

// Register adds a route. Panics if prefix is empty or handler is nil.
func (d *Dispatcher) Register(prefix string, h Handler) {
	if prefix == "" {
		panic("dispatcher: prefix must not be empty")
	}
	if h == nil {
		panic("dispatcher: handler must not be nil")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.routes = append(d.routes, Route{Prefix: prefix, Handler: h})
}

// Deregister removes all routes whose prefix equals the given value.
func (d *Dispatcher) Deregister(prefix string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	filtered := d.routes[:0]
	for _, r := range d.routes {
		if r.Prefix != prefix {
			filtered = append(filtered, r)
		}
	}
	d.routes = filtered
}

// Dispatch sends event to all matching handlers. If no handler matches and a
// fallback is configured, the fallback is called.
func (d *Dispatcher) Dispatch(event watcher.Event) {
	d.mu.RLock()
	routes := make([]Route, len(d.routes))
	copy(routes, d.routes)
	d.mu.RUnlock()

	matched := false
	for _, r := range routes {
		if strings.HasPrefix(event.Path, r.Prefix) {
			r.Handler(event)
			matched = true
		}
	}
	if !matched && d.fallback != nil {
		d.fallback(event)
	}
}

// String returns a human-readable summary of registered routes.
func (d *Dispatcher) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var sb strings.Builder
	for i, r := range d.routes {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%q", r.Prefix))
	}
	return fmt.Sprintf("Dispatcher{routes:[%s]}", sb.String())
}

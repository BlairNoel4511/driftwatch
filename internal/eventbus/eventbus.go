// Package eventbus provides a simple publish/subscribe mechanism for
// routing drift events to multiple independent consumers.
package eventbus

import (
	"context"
	"sync"
)

// Event represents a drift event published on the bus.
type Event struct {
	Path    string
	Kind    string // e.g. "drift", "missing", "restored"
	Payload any
}

// Handler is a function that receives an Event.
type Handler func(Event)

// Bus is a thread-safe publish/subscribe event bus.
type Bus struct {
	mu       sync.RWMutex
	subs     map[string][]Handler
	bufSize  int
}

// New returns a new Bus. bufSize controls the channel buffer per subscriber
// when publishing asynchronously; use 0 for synchronous dispatch.
func New(bufSize int) *Bus {
	if bufSize < 0 {
		panic("eventbus: bufSize must be >= 0")
	}
	return &Bus{
		subs:    make(map[string][]Handler),
		bufSize: bufSize,
	}
}

// Subscribe registers h to receive events of the given topic.
// Returns an unsubscribe function.
func (b *Bus) Subscribe(topic string, h Handler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[topic] = append(b.subs[topic], h)
	idx := len(b.subs[topic]) - 1
	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		slice := b.subs[topic]
		if idx < len(slice) {
			b.subs[topic] = append(slice[:idx], slice[idx+1:]...)
		}
	}
}

// Publish sends e to all subscribers of e's topic.
// If ctx is cancelled before delivery the call returns early.
func (b *Bus) Publish(ctx context.Context, e Event) {
	b.mu.RLock()
	handlers := make([]Handler, len(b.subs[e.Kind]))
	copy(handlers, b.subs[e.Kind])
	b.mu.RUnlock()

	for _, h := range handlers {
		select {
		case <-ctx.Done():
			return
		default:
			h(e)
		}
	}
}

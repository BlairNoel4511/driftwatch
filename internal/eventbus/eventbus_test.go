package eventbus_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"driftwatch/internal/eventbus"
)

func TestSubscribe_ReceivesPublishedEvent(t *testing.T) {
	b := eventbus.New(0)
	var got eventbus.Event
	b.Subscribe("drift", func(e eventbus.Event) { got = e })

	e := eventbus.Event{Path: "/etc/hosts", Kind: "drift", Payload: "size"}
	b.Publish(context.Background(), e)

	if got.Path != e.Path {
		t.Fatalf("expected path %q, got %q", e.Path, got.Path)
	}
}

func TestSubscribe_MultipleHandlers(t *testing.T) {
	b := eventbus.New(0)
	var mu sync.Mutex
	var calls []string

	record := func(id string) eventbus.Handler {
		return func(e eventbus.Event) {
			mu.Lock()
			calls = append(calls, id)
			mu.Unlock()
		}
	}
	b.Subscribe("drift", record("a"))
	b.Subscribe("drift", record("b"))

	b.Publish(context.Background(), eventbus.Event{Kind: "drift"})

	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
}

func TestUnsubscribe_StopsDelivery(t *testing.T) {
	b := eventbus.New(0)
	var count int
	unsub := b.Subscribe("drift", func(e eventbus.Event) { count++ })

	b.Publish(context.Background(), eventbus.Event{Kind: "drift"})
	unsub()
	b.Publish(context.Background(), eventbus.Event{Kind: "drift"})

	if count != 1 {
		t.Fatalf("expected 1 delivery, got %d", count)
	}
}

func TestPublish_DifferentTopicsIsolated(t *testing.T) {
	b := eventbus.New(0)
	var driftCount, missingCount int
	b.Subscribe("drift", func(e eventbus.Event) { driftCount++ })
	b.Subscribe("missing", func(e eventbus.Event) { missingCount++ })

	b.Publish(context.Background(), eventbus.Event{Kind: "drift"})
	b.Publish(context.Background(), eventbus.Event{Kind: "drift"})
	b.Publish(context.Background(), eventbus.Event{Kind: "missing"})

	if driftCount != 2 || missingCount != 1 {
		t.Fatalf("drift=%d missing=%d, want 2/1", driftCount, missingCount)
	}
}

func TestPublish_ContextCancelledStopsDispatch(t *testing.T) {
	b := eventbus.New(0)
	var count int

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	cancel() // cancel immediately

	b.Subscribe("drift", func(e eventbus.Event) { count++ })
	b.Publish(ctx, eventbus.Event{Kind: "drift"})

	if count != 0 {
		t.Fatalf("expected 0 deliveries after cancel, got %d", count)
	}
}

func TestNew_PanicsOnNegativeBufSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for negative bufSize")
		}
	}()
	eventbus.New(-1)
}

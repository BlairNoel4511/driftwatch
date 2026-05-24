package limiter_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/limiter"
)

func TestNew_PanicsOnZeroConcurrency(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero maxConcurrent")
		}
	}()
	limiter.New(0)
}

func TestAcquire_And_Release_Basic(t *testing.T) {
	l := limiter.New(2)
	ctx := context.Background()

	if err := l.Acquire(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := l.InFlight(); got != 1 {
		t.Fatalf("expected 1 in-flight, got %d", got)
	}
	l.Release()
	if got := l.InFlight(); got != 0 {
		t.Fatalf("expected 0 in-flight after release, got %d", got)
	}
}

func TestAcquire_BlocksAtCap(t *testing.T) {
	l := limiter.New(1)
	ctx := context.Background()

	if err := l.Acquire(ctx); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := l.Acquire(ctxTimeout)
	if err == nil {
		t.Fatal("expected error when cap is reached and context times out")
	}
}

func TestAcquire_ContextCancelled(t *testing.T) {
	l := limiter.New(1)
	ctx, cancel := context.WithCancel(context.Background())

	if err := l.Acquire(ctx); err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	cancel()
	if err := l.Acquire(ctx); err == nil {
		t.Fatal("expected context cancelled error")
	}
}

func TestCap_ReturnsConfiguredValue(t *testing.T) {
	l := limiter.New(5)
	if got := l.Cap(); got != 5 {
		t.Fatalf("expected cap 5, got %d", got)
	}
}

func TestLimiter_ConcurrentAcquire(t *testing.T) {
	const cap = 4
	const goroutines = 20
	l := limiter.New(cap)
	ctx := context.Background()

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		peak    int
		current int
	)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.Acquire(ctx); err != nil {
				return
			}
			mu.Lock()
			current++
			if current > peak {
				peak = current
			}
			mu.Unlock()
			time.Sleep(5 * time.Millisecond)
			mu.Lock()
			current--
			mu.Unlock()
			l.Release()
		}()
	}
	wg.Wait()

	if peak > cap {
		t.Fatalf("peak concurrency %d exceeded cap %d", peak, cap)
	}
}

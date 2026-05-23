package deadletter_test

import (
	"sync"
	"testing"

	"github.com/driftwatch/driftwatch/internal/deadletter"
)

func TestDeadLetter_ConcurrentPush(t *testing.T) {
	const workers = 20
	const perWorker = 50

	q := deadletter.New(workers * perWorker)

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				q.Push(makeEvent("/concurrent"), "err", j)
			}
		}()
	}
	wg.Wait()

	if q.Len() != workers*perWorker {
		t.Fatalf("expected %d entries, got %d", workers*perWorker, q.Len())
	}
}

func TestDeadLetter_EvictionUnderConcurrentLoad(t *testing.T) {
	const cap = 10
	q := deadletter.New(cap)

	var wg sync.WaitGroup
	wg.Add(4)
	for i := 0; i < 4; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				q.Push(makeEvent("/stress"), "err", j)
			}
		}()
	}
	wg.Wait()

	if q.Len() > cap {
		t.Fatalf("queue exceeded capacity: len=%d cap=%d", q.Len(), cap)
	}
}

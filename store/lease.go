package store

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"
)

type Lease struct {
	lock         sync.Mutex
	heap         leaseKeyHeap
	timerContext context.Context
}

func NewLease() *Lease {
	l := Lease{
		heap: *newLeaseKeyHeap(),
	}

	return &l
}

func waitAndRun(ctx context.Context, dur time.Duration, f func()) {
	select {
	case <-ctx.Done():
		fmt.Println("Context cancelled, exiting waitAndRun")
		return
	case <-time.After(dur):
		f()
	}
}

func (l *Lease) AddKey(key string, ttl time.Duration) {
	lk := leaseKey{
		key:        key,
		ttl:        ttl,
		nextExpiry: time.Now().Add(ttl),
	}

	l.lock.Lock()
	defer l.lock.Unlock()
	if len(l.heap) > 0 {
		firstKey := l.heap[0].key
		heap.Push(&l.heap, lk)
		currentKey := l.heap[0].key

		if firstKey != currentKey {
			// Cancel current timer and start a new one

		}

	} else {
		heap.Push(&l.heap, lk)

		// Create a timer for this one
		ctx := context.WithCancel(context.Background())
	}
}

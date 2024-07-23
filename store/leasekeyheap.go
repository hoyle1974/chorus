package store

import (
	"container/heap"
	"time"
)

type leaseKey struct {
	key        string
	nextExpiry time.Time
	ttl        time.Duration
}

type leaseKeyHeap []leaseKey

func (h leaseKeyHeap) Len() int           { return len(h) }
func (h leaseKeyHeap) Less(i, j int) bool { return h[i].nextExpiry.Compare(h[j].nextExpiry) < 0 }
func (h leaseKeyHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *leaseKeyHeap) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(leaseKey))
}

func (h *leaseKeyHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func newLeaseKeyHeap() *leaseKeyHeap {
	lkh := &leaseKeyHeap{}
	heap.Init(lkh)

	return lkh
}

package store

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Lease struct {
	lock   sync.RWMutex
	keys   []string
	TTL    time.Duration
	cancel context.CancelFunc
}

func NewLease(ttl time.Duration) *Lease {
	fmt.Println("NewLease", ttl)
	ctx, cancel := context.WithCancel(context.Background())
	l := &Lease{
		keys:   []string{},
		TTL:    ttl,
		cancel: cancel,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Context cancelled")
				return
			case <-time.After(ttl):
				fmt.Println("Renew")
				l.renew()
			}
		}
	}()

	return l
}

func (l *Lease) Destroy() {
	l.cancel()
}

func (l *Lease) AddKey(keys ...string) {
	fmt.Println("Adding keys", keys)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.keys = append(l.keys, keys...)
}

func (l *Lease) renew() {
	conn := getConn()
	ttl := l.TTL * 3 / 2
	l.lock.RLock()
	for _, key := range l.keys {
		fmt.Println("Renew", key, ttl)
		conn.Expire(context.Background(), key, ttl)
	}
	defer l.lock.RUnlock()
}

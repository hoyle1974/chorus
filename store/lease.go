package store

import (
	"context"
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
				// fmt.Println("Context cancelled")
				return
			case <-time.After(ttl):
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
	l.lock.Lock()
	defer l.lock.Unlock()
	l.keys = append(l.keys, keys...)
}

func (l *Lease) renew() {
	conn := getConn()
	l.lock.RLock()
	for _, key := range l.keys {
		conn.Expire(context.Background(), key, l.TTL*3/2)
	}
	defer l.lock.RUnlock()
}

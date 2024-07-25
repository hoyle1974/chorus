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

	conn := getConn()
	l.lock.RLock()
	for _, key := range l.keys {
		conn.Del(context.Background(), key)
	}
	defer l.lock.RUnlock()
}

func (l *Lease) AddKey(keys ...string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.keys = append(l.keys, keys...)
}

func (l *Lease) renew() {
	conn := getConn()
	ttl := l.TTL * 3 / 2
	l.lock.RLock()
	for _, key := range l.keys {
		conn.Expire(context.Background(), key, ttl)
	}
	defer l.lock.RUnlock()
}

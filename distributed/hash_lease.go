package distributed

import (
	"context"
	"sync"
	"time"
)

type HashLease struct {
	lock   sync.RWMutex
	hash   Hash
	keys   []string
	TTL    time.Duration
	cancel context.CancelFunc
}

func (h Hash) NewHashLease(ttl time.Duration) *HashLease {
	ctx, cancel := context.WithCancel(context.Background())
	l := &HashLease{
		hash:   h,
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

func (l *HashLease) Destroy() {
	l.cancel()

	l.lock.RLock()
	l.hash.HDel(l.keys...)
	l.keys = []string{}
	defer l.lock.RUnlock()
}

func (l *HashLease) AddKey(keys ...string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.keys = append(l.keys, keys...)
}

func (l *HashLease) renew() {
	ttl := l.TTL * 3 / 2
	l.lock.RLock()
	for _, key := range l.keys {
		// TODO fix this
		l.hash.HExpire(key, ttl)
	}
	defer l.lock.RUnlock()
}

package distributed

import (
	"context"
	"time"

	"github.com/hoyle1974/chorus/store"
	"github.com/redis/go-redis/v9"
)

type Dist struct {
	conn *redis.Client
}

func NewDist(conn *redis.Client) Dist {
	return Dist{
		conn: conn,
	}
}

func (d Dist) commonOps(key string, op ...any) time.Duration {
	ttl := time.Duration(0)
	for _, i := range op {
		switch v := i.(type) {
		case *store.Lease:
			v.AddKey(key)
		case time.Duration:
			ttl = v
		}
	}
	return ttl
}

func (d Dist) Put(key string, value string, ops ...interface{}) (string, error) {
	ttl := d.commonOps(key, ops)

	return statusCmdWrap(d.conn.Set(context.Background(), key, value, ttl))
}

func (d Dist) Get(key string) (string, error) {
	return stringCmdWrap(d.conn.Get(context.Background(), key))
}

func (d Dist) Expire(key string, ttl time.Duration) (bool, error) {
	return boolCmdWrap(d.conn.Expire(context.Background(), key, ttl))
}

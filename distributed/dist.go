package distributed

import (
	"context"
	"fmt"
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

func (d Dist) commonOps(key string, op []any) time.Duration {
	fmt.Println("commonOps", key, op)
	ttl := time.Duration(0)
	fmt.Println("	", len(op))
	for _, i := range op {
		switch v := i.(type) {
		case *store.Lease:
			v.AddKey(key)
			ttl = v.TTL
		case time.Duration:
			ttl = v
		}
	}
	return ttl
}

func (d Dist) Put(key string, value string, ops ...any) (string, error) {
	ttl := d.commonOps(key, ops)

	return statusCmdWrap(d.conn.Set(context.Background(), key, value, ttl))
}

func (d Dist) PutIfAbsent(key string, value string, ops ...any) (bool, error) {
	ttl := d.commonOps(key, ops)
	return boolCmdWrap(d.conn.SetNX(context.Background(), key, value, ttl))
}

func (d Dist) Get(key string) (string, error) {
	return stringCmdWrap(d.conn.Get(context.Background(), key))
}

func (d Dist) Expire(key string, ttl time.Duration) (bool, error) {
	return boolCmdWrap(d.conn.Expire(context.Background(), key, ttl))
}

package ds

import (
	"sync/atomic"

	"github.com/redis/go-redis/v9"
)

var conn atomic.Pointer[redis.Client]

func GetConn() *redis.Client {
	if conn.Load() != nil {
		return conn.Load()
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	old := conn.Swap(rdb)
	if old != nil {
		old.Close()
	}
	return rdb
}

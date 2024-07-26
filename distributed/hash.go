package distributed

import (
	"context"
	"time"
)

// Commands we support:
// HDel, HExists, HExpire,HExpireAt,HExpireTime, HGet,HGetAll,HIncrBy,HIncrByFloat,HKeys,
// HLen, HMGet, HMSet, HPersist, HPExpire, HPExpireAt, HPExpireTime, HPTTL, HRandField, HScan
// Hset, HSetNX, HStrLen, HTTL, HVals
type Hash struct {
	dist Dist
	key  string
}

func (d Dist) BindHash(key string, op ...any) Hash {
	d.commonOps(key, op)
	return Hash{key: key, dist: d}
}

func (h Hash) HDel(keys ...string) (int64, error) {
	return intCmdWrap(h.dist.conn.HDel(context.Background(), h.key, keys...))
}

func (h Hash) HExists(key string) (bool, error) {
	return boolCmdWrap(h.dist.conn.HExists(context.Background(), h.key, key))
}

func (h Hash) HExpire(key string, ttl time.Duration) ([]int64, error) {
	return intSliceCmdWrap(h.dist.conn.HExpire(context.Background(), h.key, ttl, key))
}

func (h Hash) HGet(key string) (string, error) {
	return stringCmdWrap(h.dist.conn.HGet(context.Background(), h.key, key))
}

func (h Hash) HSet(key string, value string) (int64, error) {
	return intCmdWrap(h.dist.conn.HSet(context.Background(), h.key, key, value))
}

func (h Hash) HSetNX(key string, value string) (bool, error) {
	return boolCmdWrap(h.dist.conn.HSetNX(context.Background(), h.key, key, value))
}

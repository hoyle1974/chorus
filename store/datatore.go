package store

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/hoyle1974/chorus/misc"
	"github.com/redis/go-redis/v9"
)

var conn atomic.Pointer[redis.Client]

func getConn() *redis.Client {
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

func PutIfAbsent(key string, value string, ttl time.Duration) (bool, error) {
	bc := getConn().SetNX(context.Background(), key, value, ttl)
	return bc.Val(), bc.Err()
}

func Put(key string, value string, ttl time.Duration) error {
	return getConn().Set(context.Background(), key, value, ttl).Err()
}

func Expire(key string, ttl time.Duration) error {
	return getConn().Expire(context.Background(), key, ttl).Err()
}

func Get(key string) (string, error) {
	sc := getConn().Get(context.Background(), key)
	return sc.Val(), sc.Err()
}

func Del(key string) error {
	return getConn().Del(context.Background(), key).Err()
}

func GetSet(key string) ([]string, error) {
	sc := getConn().SMembers(context.Background(), key)
	return sc.Val(), sc.Err()
}

func AddMemberToSet(key string, members ...string) error {
	return getConn().SAdd(context.Background(), key, members).Err()
}

func RemoveMemberFromSet(key string, members ...string) error {
	return getConn().SRem(context.Background(), key, members).Err()
}

// ------------------ composite commands

func PutConnectionInfo(machineId misc.MachineId, id misc.ConnectionId) {
	key := "connections/" + string(id)
	Put(key, string(machineId), time.Duration(0))
}
func RemoveConnectionInfo(machineId misc.MachineId, id misc.ConnectionId) {
	key := "connections/" + string(id)
	Del(key)
}
func GetConnectionInfo(id misc.ConnectionId) misc.MachineId {
	key := "connections/" + string(id)
	m, _ := Get(key)
	return misc.MachineId(m)
}

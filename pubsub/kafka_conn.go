package pubsub

import (
	"sync/atomic"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

var conn atomic.Pointer[kgo.Client]
var brokers = []string{"localhost:19092"}

func newConn() (*kgo.Client, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)

	return client, err
}

func getConn() *kgo.Client {
	if conn.Load() != nil {
		return conn.Load()
	}

	db, err := newConn()
	if err != nil {
		panic(err)
	}

	old := conn.Swap(db)
	if old != nil {
		old.Close()
	}
	return db
}

func getAdminConn() *kadm.Client {
	return kadm.NewClient(getConn())
}

package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/hoyle1974/chorus/ds"
	"github.com/hoyle1974/chorus/misc"
	"github.com/redis/go-redis/v9"
)

type TopicMessageHandler interface {
	OnMessageFromTopic(msg Message)
}

type Message interface {
	String() string
	Topic() misc.TopicId
	Unmarshal([]byte)
}

func SendMessage(msg Message) {
	fmt.Println("pubsub.SendMessage to", msg.Topic(), ":", msg)
	err := ds.GetConn().Publish(context.Background(), string(msg.Topic()), msg.String()).Err()
	if err != nil {
		panic(err)
	}
}

type Consumer struct {
	log        *slog.Logger
	topic      misc.TopicId
	msgHandler TopicMessageHandler
	pubsub     *redis.PubSub
	ready      atomic.Bool
}

func NewConsumer(log *slog.Logger, topic misc.TopicId, msgHandler TopicMessageHandler) *Consumer {
	newLog := log.With("topic", topic)

	newLog.Debug("pubsub.NewConsumer", "topic", topic)
	pubsub := ds.GetConn().PSubscribe(context.Background(), string(topic))
	consumer := &Consumer{topic: topic, msgHandler: msgHandler, pubsub: pubsub}
	go func() {
		// If we haven't start this consumer in 10 seconds then log something
		time.Sleep(time.Duration(10) * time.Second)
		if !consumer.ready.Load() {
			newLog.Error("Created a consumer for a topicbut it was not started within 10 seconds!")
		}
	}()

	return consumer
}

func (c *Consumer) StartConsumer(v Message) {
	c.ready.Store(true)
	go c.processMessages(v)
}

func (c *Consumer) processMessages(v Message) {
	// Listen for messages
	ch := c.pubsub.Channel()
	for redisMsg := range ch {
		v.Unmarshal([]byte(redisMsg.Payload))
		c.msgHandler.OnMessageFromTopic(v)
	}
}

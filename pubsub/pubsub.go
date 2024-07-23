package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hoyle1974/chorus/ds"
	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/redis/go-redis/v9"
)

type TopicMessageHandler interface {
	OnMessageFromTopic(msg message.Message)
}

func SendMessage(msg message.Message) {
	fmt.Println("pubsub.SendMessage to", msg.RoomId.Topic(), ":", msg)
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	err = ds.GetConn().Publish(context.Background(), string(msg.RoomId.Topic()), b).Err()
	if err != nil {
		panic(err)
	}
}

type Consumer struct {
	topic      misc.TopicId
	msgHandler TopicMessageHandler
	pubsub     *redis.PubSub
}

func NewConsumer(topic misc.TopicId, msgHandler TopicMessageHandler) *Consumer {
	pubsub := ds.GetConn().PSubscribe(context.Background(), string(topic))

	return &Consumer{topic: topic, msgHandler: msgHandler, pubsub: pubsub}
}

func (c *Consumer) StartConsumer() {
	go c.processMessages()
}

func (c *Consumer) processMessages() {
	fmt.Println("pubsub.Consumer.processMessages", c.topic)
	// Listen for messages
	ch := c.pubsub.Channel()
	for rmsg := range ch {
		var msg message.Message

		if err := json.Unmarshal([]byte(rmsg.Payload), &msg); err != nil {
			fmt.Println("--- error ", err)
		} else {
			c.msgHandler.OnMessageFromTopic(msg)
			fmt.Println("pubsub.Consumer.ProcessMessage from", c.topic, ":", msg)
		}
	}
}

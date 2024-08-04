package pubsub

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/hoyle1974/chorus/misc"
	"github.com/twmb/franz-go/pkg/kgo"
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
	getConn().Produce(
		context.Background(),
		&kgo.Record{
			Topic: string(msg.Topic()),
			Value: []byte(msg.String()),
		}, nil)
}

type Consumer struct {
	log        *slog.Logger
	topic      misc.TopicId
	msgHandler TopicMessageHandler
	pubsub     *kgo.Client
	ready      atomic.Bool
}

func TopicExists(topic misc.TopicId) bool {
	ctx := context.Background()
	client := getAdminConn()
	defer client.Close()

	topicsMetadata, err := client.ListTopics(ctx)
	if err != nil {
		panic(err)
	}
	for _, metadata := range topicsMetadata {
		if metadata.Topic == string(topic) {
			return true
		}
	}
	return false
}

func CreateTopic(topic misc.TopicId) {
	ctx := context.Background()
	client := getAdminConn()
	defer client.Close()

	_, err := client.CreateTopic(ctx, 1, 1, nil, string(topic))
	if err != nil {
		panic(err)
	}
}

func DeleteTopic(topic misc.TopicId) {
	ctx := context.Background()
	client := getAdminConn()
	defer client.Close()

	_, err := client.DeleteTopic(ctx, string(topic))
	if err != nil {
		panic(err)
	}
}

func NewConsumer(log *slog.Logger, groupID string, topic misc.TopicId, msgHandler TopicMessageHandler) *Consumer {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(string(topic)),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
	)
	if err != nil {
		log.Error("Error creating client", "error", err)
		return nil
	}

	consumer := &Consumer{log: log, topic: topic, msgHandler: msgHandler, pubsub: client}

	return consumer
}

func (c *Consumer) AddTopic(topic misc.TopicId) {
	c.pubsub.AddConsumeTopics(string(topic))
}

func (c *Consumer) RemoveTopic(topic misc.TopicId) {
	c.log.Warn("RemoveTopic - not implemented")
}

func (c *Consumer) StartConsumer(v Message) {
	c.ready.Store(true)
	go c.processMessages(v)
}

func (c *Consumer) processMessages(v Message) {
	// Listen for messages
	ctx := context.Background()
	for {
		fetches := c.pubsub.PollFetches(ctx)
		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			v.Unmarshal([]byte(record.Value))
			c.msgHandler.OnMessageFromTopic(v)
		}
	}
}

package pubsub

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"

	"fmt"
)

var admin = newAdmin()

func TopicExists(topic string) bool {
	return admin.TopicExists(topic)
}
func CreateTopic(topic string) error {
	return admin.CreateTopic(topic)
}
func DeleteTopic(topic string) error {
	return admin.DeleteTopic(topic)
}

var producerLock = sync.RWMutex{}
var producers = map[misc.RoomId]*Producer{}

func SendMessage(msg message.Message) {
	producerLock.RLock()
	producer, ok := producers[msg.RoomId]
	producerLock.RUnlock()
	if ok {
		producer.SendMessage(msg)
		return
	}

	producerLock.Lock()
	producer, ok = producers[msg.RoomId]
	if !ok {
		producer = newProducer(string(msg.RoomId))
		producers[msg.RoomId] = producer
	}
	producerLock.Unlock()
	producer.SendMessage(msg)
}

type Admin struct {
	client *kadm.Client
}

var brokers = []string{
	"localhost:19092",
}

func newAdmin() *Admin {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)
	if err != nil {
		panic(err)
	}
	admin := kadm.NewClient(client)
	return &Admin{client: admin}
}
func (a *Admin) TopicExists(topic string) bool {
	ctx := context.Background()
	topicsMetadata, err := a.client.ListTopics(ctx)
	if err != nil {
		panic(err)
	}
	for _, metadata := range topicsMetadata {
		if metadata.Topic == topic {
			return true
		}
	}
	return false
}
func (a *Admin) CreateTopic(topic string) error {
	ctx := context.Background()

	configs := map[string]*string{}
	val := "1"
	configs["retention.ms"] = &val
	resp, err := a.client.CreateTopics(ctx, 1, 1, configs, topic)
	if err != nil {
		panic(err)
	}
	for _, ctr := range resp {
		if ctr.Err != nil {
			return fmt.Errorf("unable to create topic '%s': %w", ctr.Topic, ctr.Err)
		}
	}
	return nil
}
func (a *Admin) DeleteTopic(topic string) error {
	ctx := context.Background()

	resp, err := a.client.DeleteTopic(ctx, topic)
	if err != nil {
		panic(err)
	}
	if resp.Err != nil {
		fmt.Errorf("unable to delete topic '%s': %w", resp.Topic, resp.Err)
	}

	return nil
}
func (a *Admin) Close() {
	a.client.Close()
}

type Producer struct {
	client *kgo.Client
	topic  string
}

func newProducer(topic string) *Producer {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)
	if err != nil {
		panic(err)
	}
	return &Producer{client: client, topic: topic}
}
func (p *Producer) SendMessage(msg message.Message) {
	ctx := context.Background()
	b, _ := json.Marshal(msg)
	p.client.Produce(ctx, &kgo.Record{Topic: p.topic, Value: b}, func(r *kgo.Record, err error) {
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		}
	})
}
func (p *Producer) Close() {
	p.client.Close()
}

type TopicMessageHandler interface {
	OnMessageFromTopic(msg message.Message)
}

type Consumer struct {
	client     *kgo.Client
	topic      string
	msgHandler TopicMessageHandler
}

func NewConsumer(topic string, msgHandler TopicMessageHandler) *Consumer {
	groupID := misc.UUIDString()
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
	)
	if err != nil {
		panic(err)
	}
	return &Consumer{client: client, topic: topic, msgHandler: msgHandler}
}
func (c *Consumer) ProcessMessages() {
	ctx := context.Background()
	for {
		fetches := c.client.PollFetches(ctx)
		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			var msg message.Message
			if err := json.Unmarshal(record.Value, &msg); err != nil {
				fmt.Printf("Error decoding message: %v\n", err)
				continue
			}
			// Do something with the message
			c.msgHandler.OnMessageFromTopic(msg)
		}
	}
}
func (c *Consumer) Close() {
	c.client.Close()
}

package kafka

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"github.com/smartwalle/mx"
	"sync"
)

type Consumer struct {
	closed   bool
	mu       *sync.Mutex
	topic    string
	group    string
	client   sarama.Client
	consumer *consumer
}

func NewConsumer(topic, group string, config *Config) (*Consumer, error) {
	client, err := sarama.NewClient(config.Addrs, config.Config)
	if err != nil {
		return nil, err
	}

	var c = &Consumer{}
	c.closed = false
	c.mu = &sync.Mutex{}
	c.topic = topic
	c.group = group
	c.client = client
	return c, nil
}

func (c *Consumer) Dequeue(handler mx.Handler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return mx.ErrClosedQueue
	}

	if c.consumer != nil {
		c.consumer.Close()
		<-c.consumer.stopChan
		c.consumer = nil
	}

	if c.consumer == nil {
		consumer, err := newConsumer(c.topic, c.group, c.client, handler)
		if err != nil {
			return err
		}
		c.consumer = consumer
	}

	return nil
}

func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return mx.ErrClosedQueue
	}

	c.closed = true

	if c.consumer != nil {
		var err = c.consumer.Close()
		<-c.consumer.stopChan
		if err != nil {
			return err
		}
		c.consumer = nil
	}

	return nil
}

type consumer struct {
	mu        *sync.Mutex
	closed    bool
	readyChan chan struct{}
	stopChan  chan struct{}
	topics    []string
	cancel    context.CancelFunc
	consumer  sarama.ConsumerGroup
	handler   mx.Handler
}

func newConsumer(topic, group string, client sarama.Client, handler mx.Handler) (*consumer, error) {
	consumerGroup, err := sarama.NewConsumerGroupFromClient(group, client)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	var c = &consumer{}
	c.mu = &sync.Mutex{}
	c.closed = false
	c.readyChan = make(chan struct{})
	c.stopChan = make(chan struct{})
	c.topics = []string{topic}
	c.cancel = cancel
	c.consumer = consumerGroup
	c.handler = handler

	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if nErr := consumerGroup.Consume(ctx, c.topics, c); nErr != nil {
				if errors.Is(nErr, sarama.ErrClosedConsumerGroup) {
					return
				}
			}

			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}

			c.readyChan = make(chan struct{})
		}
	}()
	<-c.readyChan
	return c, nil
}

func (c *consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	c.cancel()
	return c.consumer.Close()
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.readyChan)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *consumer) Cleanup(sarama.ConsumerGroupSession) error {
	if c.closed {
		close(c.stopChan)
	}
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			var m = &Message{}
			m.m = message
			if c.handler(m) {
				session.MarkMessage(message, "")
			}
		case <-session.Context().Done():
			return nil
		}
	}
	return nil
}

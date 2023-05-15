package pulsar

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/smartwalle/mx"
	"sync/atomic"
	"time"
)

type Producer struct {
	closed   int32
	topic    string
	config   *Config
	client   pulsar.Client
	producer pulsar.Producer
}

func NewProducer(topic string, config *Config) (*Producer, error) {
	client, err := pulsar.NewClient(config.ClientOptions)
	if err != nil {
		return nil, err
	}
	config.ProducerOptions.Topic = topic
	producer, err := client.CreateProducer(config.ProducerOptions)
	if err != nil {
		return nil, err
	}

	var p = &Producer{}
	p.closed = 0
	p.topic = topic
	p.config = config
	p.producer = producer
	return p, nil
}

func (this *Producer) Enqueue(data []byte) error {
	var m = NewProducerMessage()
	m.Payload = data
	return this.EnqueueMessage(m)
}

func (this *Producer) EnqueueMessage(m *pulsar.ProducerMessage) error {
	if m == nil {
		return nil
	}

	if atomic.LoadInt32(&this.closed) == 1 {
		return mx.ErrClosedQueue
	}

	_, err := this.producer.Send(context.Background(), m)
	return err
}

func (this *Producer) DeferredEnqueue(delay time.Duration, data []byte) error {
	var m = NewProducerMessage()
	m.Payload = data
	m.DeliverAfter = delay
	return this.EnqueueMessage(m)
}

func (this *Producer) MultiEnqueue(data ...[]byte) error {
	if len(data) == 0 {
		return nil
	}

	if atomic.LoadInt32(&this.closed) == 1 {
		return mx.ErrClosedQueue
	}

	for _, d := range data {
		var m = NewProducerMessage()
		m.Payload = d
		if _, err := this.producer.Send(context.Background(), m); err != nil {
			return err
		}
	}
	return nil
}

func (this *Producer) Close() error {
	if !atomic.CompareAndSwapInt32(&this.closed, 0, 1) {
		return nil
	}

	if this.producer != nil {
		this.producer.Close()
		this.producer = nil
	}

	if this.client != nil {
		this.client.Close()
		this.client = nil
	}

	return nil
}

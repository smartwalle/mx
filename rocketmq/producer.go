package rocketmq

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/smartwalle/mx"
	"sync"
)

type Producer struct {
	mu       *sync.Mutex
	closed   bool
	config   *Config
	producer rocketmq.Producer
}

func NewProducer(config *Config) (*Producer, error) {
	var opts []producer.Option
	opts = append(opts, producer.WithGroupName(config.Producer.Group))
	opts = append(opts, producer.WithInstanceName(config.InstanceName))
	opts = append(opts, producer.WithNameServer(config.NameServerAddrs))
	opts = append(opts, producer.WithNameServerDomain(config.NameServerDomain))
	opts = append(opts, producer.WithNamespace(config.Namespace))
	opts = append(opts, producer.WithVIPChannel(config.VIPChannelEnabled))
	opts = append(opts, producer.WithRetry(config.RetryTimes))
	opts = append(opts, producer.WithCredentials(config.Credentials))

	opts = append(opts, producer.WithInterceptor(config.Producer.Interceptors...))
	opts = append(opts, producer.WithSendMsgTimeout(config.Producer.SendMsgTimeout))
	opts = append(opts, producer.WithQueueSelector(config.Producer.Selector))
	opts = append(opts, producer.WithDefaultTopicQueueNums(config.Producer.DefaultTopicQueueNums))
	opts = append(opts, producer.WithCreateTopicKey(config.Producer.CreateTopicKey))

	var producer, err = rocketmq.NewProducer(opts...)
	if err != nil {
		return nil, err
	}

	if err = producer.Start(); err != nil {
		return nil, err
	}

	var p = &Producer{}
	p.mu = &sync.Mutex{}
	p.closed = false
	p.config = config
	p.producer = producer
	return p, nil
}

func (this *Producer) Enqueue(topic string, data []byte) error {
	var m = primitive.NewMessage(topic, data)
	return this.EnqueueMessage(m)
}

func (this *Producer) EnqueueMessage(m *primitive.Message) error {
	if m == nil {
		return nil
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	if this.closed {
		return mx.ErrClosedQueue
	}

	_, err := this.producer.SendSync(context.Background(), m)
	return err
}

func (this *Producer) Close() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.closed {
		return nil
	}
	this.closed = true

	if this.producer != nil {
		if err := this.producer.Shutdown(); err != nil {
			return err
		}
		this.producer = nil
	}

	return nil
}

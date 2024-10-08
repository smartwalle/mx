package rocketmq

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/smartwalle/mx"
	"sync/atomic"
)

type Producer struct {
	closed   int32
	topic    string
	config   *Config
	producer rocketmq.Producer
}

func NewProducer(topic string, config *Config) (*Producer, error) {
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
	p.closed = 0
	p.topic = topic
	p.config = config
	p.producer = producer
	return p, nil
}

func (p *Producer) Enqueue(data []byte) error {
	var m = primitive.NewMessage(p.topic, data)
	return p.EnqueueMessages(m)
}

func (p *Producer) EnqueueMessage(m *primitive.Message) error {
	m.Topic = p.topic
	return p.EnqueueMessages(m)
}

func (p *Producer) MultiEnqueue(data ...[]byte) error {
	if len(data) == 0 {
		return nil
	}

	var ms = make([]*primitive.Message, 0, len(data))
	for _, d := range data {
		var m = primitive.NewMessage(p.topic, d)
		ms = append(ms, m)
	}
	return p.EnqueueMessages(ms...)
}

func (p *Producer) EnqueueMessages(m ...*primitive.Message) error {
	if len(m) == 0 {
		return nil
	}

	if atomic.LoadInt32(&p.closed) == 1 {
		return mx.ErrClosedQueue
	}

	_, err := p.producer.SendSync(context.Background(), m...)
	return err
}

func (p *Producer) Close() error {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return nil
	}

	if p.producer != nil {
		if err := p.producer.Shutdown(); err != nil {
			return err
		}
	}

	return nil
}

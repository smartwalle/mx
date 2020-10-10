package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/smartwalle/mx"
	"sync"
)

type Config struct {
	*sarama.Config
	Addrs []string
}

func NewConfig() *Config {
	var c = &Config{}
	c.Config = sarama.NewConfig()
	c.Addrs = []string{"127.0.0.1:9092"}
	c.Version = sarama.V2_1_0_0

	// 等待服务器所有副本都保存成功后的响应
	c.Producer.RequiredAcks = sarama.WaitForAll
	// 随机的分区类型：返回一个分区器，该分区器每次选择一个随机分区
	c.Producer.Partitioner = sarama.NewRandomPartitioner
	// 是否等待成功和失败后的响应
	c.Config.Producer.Return.Successes = true
	c.Config.Producer.Return.Errors = true

	c.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	c.Consumer.Offsets.Initial = sarama.OffsetOldest
	return c
}

type Queue struct {
	mu            *sync.Mutex
	closed        bool
	topic         string
	client        sarama.Client
	producer      sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
	consumer      *consumer
}

func New(topic string, config *Config) (*Queue, error) {
	client, err := sarama.NewClient(config.Addrs, config.Config)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	asyncProducer, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}

	var q = &Queue{}
	q.mu = &sync.Mutex{}
	q.closed = false
	q.topic = topic
	q.client = client
	q.producer = producer
	q.asyncProducer = asyncProducer
	return q, nil
}

func (this *Queue) Enqueue(data []byte) error {
	var m = &sarama.ProducerMessage{}
	m.Topic = this.topic
	//m.Partition =
	//m.Key =
	m.Value = sarama.ByteEncoder(data)
	return this.EnqueueMessage(m)
}

func (this *Queue) EnqueueMessage(m *sarama.ProducerMessage) error {
	if m == nil {
		return nil
	}

	if this.closed {
		return mx.ErrClosedQueue
	}

	_, _, err := this.producer.SendMessage(m)
	return err
}

//func (this *Queue) AsyncEnqueue(value []byte, h func(error)) {
//	var m = &sarama.ProducerMessage{}
//	m.Topic = this.topic
//	//m.Partition =
//	//m.Key =
//	m.Value = sarama.ByteEncoder(value)
//	this.AsyncEnqueueMessage(m, h)
//}
//
//func (this *Queue) AsyncEnqueueMessage(m *sarama.ProducerMessage, h func(error)) {
//	if m == nil {
//		return
//	}
//
//	if this.closed {
//		return
//	}
//
//	this.asyncProducer.Input() <- m
//
//	select {
//	case <-this.asyncProducer.Successes():
//		if h != nil {
//			h(nil)
//		}
//	case err := <-this.asyncProducer.Errors():
//		if h != nil {
//			h(err)
//		}
//	}
//}

func (this *Queue) Dequeue(group string, handler mx.Handler) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.closed {
		return mx.ErrClosedQueue
	}

	if this.consumer != nil {
		this.consumer.Close()
		<-this.consumer.stopChan
		this.consumer = nil
	}

	if this.consumer == nil {
		consumer, err := newConsumer(this.topic, group, this.client, handler)
		if err != nil {
			return err
		}
		this.consumer = consumer
	}

	return nil
}

func (this *Queue) Close() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.closed {
		return nil
	}
	this.closed = true

	if this.consumer != nil {
		var err = this.consumer.Close()
		<-this.consumer.stopChan
		if err != nil {
			return err
		}
	}

	if this.producer != nil {
		var err = this.producer.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

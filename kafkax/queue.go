package kafkax

import (
	"github.com/segmentio/kafka-go"
	"github.com/smartwalle/mx"
	"net"
)

type Config struct {
	Reader kafka.ReaderConfig
	Writer kafka.Writer
}

func TCP(address ...string) net.Addr {
	return kafka.TCP(address...)
}

func NewConfig() *Config {
	return &Config{}
}

type Queue struct {
	producer *Producer
	consumer *Consumer
	config   *Config
	topic    string
}

func NewQueue(topic string, config *Config) (*Queue, error) {
	producer, err := NewProducer(topic, config)
	if err != nil {
		return nil, err
	}

	var q = &Queue{}
	q.producer = producer
	q.config = config
	q.topic = topic
	return q, nil
}

func (this *Queue) Enqueue(data []byte) error {
	return this.producer.Enqueue(data)
}

func (this *Queue) MultiEnqueue(data ...[]byte) error {
	return this.producer.MultiEnqueue(data...)
}

func (this *Queue) Dequeue(group string, handler mx.Handler) error {
	if this.consumer != nil {
		this.consumer.Close()
	}

	var err error
	this.consumer, err = NewConsumer(this.topic, group, this.config)
	if err != nil {
		return err
	}
	return this.consumer.Dequeue(handler)
}

func (this *Queue) Close() error {
	if this.consumer != nil {
		if err := this.consumer.Close(); err != nil {
			return err
		}
	}

	if this.producer != nil {
		if err := this.producer.Close(); err != nil {
			return err
		}
	}
	return nil
}
package nats

import (
	n "github.com/nats-io/nats.go"
	"github.com/smartwalle/mx"
	"sync"
)

type Consumer struct {
	closed bool
	mu     *sync.Mutex
	topic  string
	group  string
	config *Config
	conn   *n.Conn
	sub    *n.Subscription
}

func NewConsumer(topic, group string, config *Config) (*Consumer, error) {
	conn, err := config.Connect()
	if err != nil {
		return nil, err
	}
	if err = conn.Flush(); err != nil {
		return nil, err
	}
	if err = conn.LastError(); err != nil {
		return nil, err
	}

	var c = &Consumer{}
	c.closed = false
	c.mu = &sync.Mutex{}
	c.topic = topic
	c.group = group
	c.config = config
	c.conn = conn
	return c, nil
}

func (c *Consumer) Dequeue(handler mx.Handler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return mx.ErrClosedQueue
	}

	if c.sub != nil {
		c.sub.Unsubscribe()
	}

	sub, err := c.conn.QueueSubscribe(c.topic, c.group, func(msg *n.Msg) {
		var m = &Message{}
		m.m = msg
		handler(m)
	})
	if err != nil {
		return err
	}
	c.sub = sub

	return nil
}

func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return mx.ErrClosedQueue
	}

	c.closed = true

	if c.sub != nil {
		c.sub.Unsubscribe()
		c.sub = nil
	}

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	return nil
}

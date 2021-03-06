package main

import (
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/smartwalle/mx"
	"github.com/smartwalle/mx/rocketmq"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var config = rocketmq.NewConfig()
	config.NameServerAddrs = []string{"192.168.1.77:9876"}
	config.Consumer.FromWhere = consumer.ConsumeFromFirstOffset
	config.Consumer.ConsumeOrderly = true
	c, err := rocketmq.NewConsumer("topic-1", "group-1", config)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.Dequeue(func(m mx.Message) bool {
		var mm = m.(*rocketmq.Message)
		fmt.Println("Dequeue", mm.Message().Queue.QueueId, time.Now(), string(mm.Value()))
		return true
	})

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig:
	}
	fmt.Println("Close", c.Close())
}

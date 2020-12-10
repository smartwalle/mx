package main

import (
	"fmt"
	"github.com/smartwalle/mx"
	"github.com/smartwalle/mx/kafka"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var config = kafka.NewConfig()
	config.Addrs = []string{"192.168.1.77:9092"}
	c, err := kafka.NewConsumer("topic-1", "group-1", config)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.Dequeue(func(m mx.Message) bool {
		fmt.Println("Dequeue 1", time.Now(), string(m.Value()))
		return true
	})
	c.Dequeue(func(m mx.Message) bool {
		fmt.Println("Dequeue 2", time.Now(), string(m.Value()))
		return true
	})

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig:
	}
	fmt.Println("Close", c.Close())
}

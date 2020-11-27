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
	c, err := kafka.NewConsumer("topic-1", "group-1", config)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.Dequeue(func(m mx.Message) bool {
		fmt.Println("Dequeue", time.Now(), string(m.Value()))
		return true
	})

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig:
	}
	fmt.Println("Close", c.Close())
}

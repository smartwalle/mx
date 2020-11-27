package main

import (
	"fmt"
	"github.com/smartwalle/mx/rocketmq"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var config = rocketmq.NewConfig()
	p, err := rocketmq.NewProducer(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("begin...")
	for i := 0; i < 1000; i++ {
		if err := p.Enqueue("topic-1", []byte(fmt.Sprintf("hello  %s %d", time.Now(), i))); err != nil {
			fmt.Println("Enqueue", err)
			break
		}
	}
	fmt.Println("end...")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig:
	}
	fmt.Println("Close", p.Close())
}

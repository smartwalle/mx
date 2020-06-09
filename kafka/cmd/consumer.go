package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/smartwalle/mx/kafka"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config := sarama.NewConfig()
	// 等待服务器所有副本都保存成功后的响应
	config.Producer.RequiredAcks = sarama.WaitForAll
	// 随机的分区类型：返回一个分区器，该分区器每次选择一个随机分区
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	// 是否等待成功和失败后的响应
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Version = sarama.V2_1_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	kc, err := sarama.NewClient([]string{"localhost:9092"}, config)
	if err != nil {
		fmt.Println(err)
		return
	}

	q, err := kafka.New("topic-1", "group-1", kc)
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		for {
			var m, err = q.Dequeue()
			if err != nil {
				fmt.Println("Dequeue", err)
				break
			}
			fmt.Println("Dequeue", string(m.Value()))
			m.Ack()
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigterm:
	}
	fmt.Println("Close", q.Close())
}

package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/smartwalle/mx/kafka"
)

// 查看 Topic 信息
// ./bin/kafka-topics.sh --describe --zookeeper 127.0.0.1 --topic topic_name
//
// 调整 Topic 分区数量
// ./bin/kafka-topics.sh --alter --zookeeper 127.0.0.1 --topic topic_name --partitions partition_count
//
// 删除 Topic
// ./bin/kafka-topics.sh --delete --zookeeper 127.0.0.1 --topic topic_name

func main() {
	var config = kafka.NewConfig()
	config.Addrs = []string{"localhost:9092"}
	// 等待服务器所有副本都保存成功后的响应
	config.Producer.RequiredAcks = sarama.WaitForAll
	// 随机的分区类型：返回一个分区器，该分区器每次选择一个随机分区
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Version = sarama.V2_1_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	q, err := kafka.New("topic-1", "group-1", config)
	if err != nil {
		fmt.Println(err)
		return
	}

	//for {
	//	if err := q.Enqueue([]byte(fmt.Sprintf("hello %s", time.Now().Format(time.RFC3339Nano)))); err != nil {
	//		fmt.Println("Enqueue", err)
	//		break
	//	}
	//}

	for i := 0; i < 100000000000; i++ {
		if err := q.Enqueue([]byte(fmt.Sprintf("hello %d", i))); err != nil {
			fmt.Println("Enqueue", err)
			break
		}
	}

	fmt.Println("end")

	select {}
}

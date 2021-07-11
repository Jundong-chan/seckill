package plugins

import (
	"encoding/json"
	"fmt"
	"github.com/Jundong-chan/seckill/model"
	"sync"

	"github.com/Shopify/sarama"
)

var brokeraddr = "localhost:9092"

//接收预订单的消息的channel
var Preordermsg = make(chan model.Order, 4096)

type KafkaConsumerService interface {
	CreateOrderConsumer()
}

type KafkaConsumerServiceImpl struct{}

func NewKafkaConsumersvcimpl() *KafkaConsumerServiceImpl {
	return &KafkaConsumerServiceImpl{}
}

//消费者取消息
func (impl KafkaConsumerServiceImpl) CreateOrderConsumer() {
	var wg sync.WaitGroup
	consumer, err := sarama.NewConsumer([]string{brokeraddr}, nil)
	if err != nil {
		fmt.Println(err)
	}
	partitionList, err := consumer.Partitions("order")
	if err != nil {
		fmt.Println(err)
	}

	for partition := range partitionList {
		pc, err := consumer.ConsumePartition("order", int32(partition), sarama.OffsetNewest)
		if err != nil {
			fmt.Println(err)
		}
		defer pc.AsyncClose()
		wg.Add(1)
		go func(sarama.PartitionConsumer) {
			defer wg.Done()
			preorder := model.Order{}
			//当读取不到会阻塞
			for msg := range pc.Messages() {
				fmt.Printf("Consumer收到消息：Partition:%d, Offset:%d, Key:%s, Value:%s\n", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
				json.Unmarshal(msg.Value, &preorder)
				Preordermsg <- preorder //反序列化后放进channel
			}
		}(pc)
		wg.Wait()
		consumer.Close()
	}
}

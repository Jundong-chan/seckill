package plugins

import (
	"github.com/Jundong-chan/seckill/model"
"encoding/json"
"fmt"
"log"
"time"

"github.com/Shopify/sarama"
)

var mas = make(chan sarama.ProducerMessage, 20000) //生成的消息放到chan里

type KafkaProducerService interface {
	AsyncProducer()
	GenerateMessage(topic string, value model.Order) *sarama.ProducerMessage
}

type KafkaProducerServiceImpl struct{}

func NewKafkaProducerimpl() *KafkaProducerServiceImpl {
	return &KafkaProducerServiceImpl{}
}

//发送消息到缓存 全局变量 mas channel中
func (kafkasvc *KafkaProducerServiceImpl) GenerateMessage(topic string, value model.Order) {
	key := "order"
	mss, err := json.Marshal(value)
	if err != nil {
		log.Println("GenerateMessage Marsal failed")
	}
	message := sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(mss),
		Key:   sarama.StringEncoder(key),
	}
	select {
	case mas <- message:
	case <-time.After(time.Second * 5):
		log.Printf("GenerateMessage overtime: %v", message)
	}

}

//生产者异步发送消息，只要将数据放到 maschannel就 ok
func (kafkasvc *KafkaProducerServiceImpl) AsyncProducer() {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	//向分区发送消息的策略，如果设置的是no response 那么return的配置无用
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Version = sarama.V0_11_0_0 //版本
	//配置broker地址
	fmt.Println("start make producer")
	defer func() {
		if err := recover(); err != nil {
			log.Printf("error: %v", err)
		}
	}()
	producer, err := sarama.NewAsyncProducer([]string{brokeraddr}, config)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer producer.Close()
	defer producer.AsyncClose()
	//接受反馈
	go func(p sarama.AsyncProducer) {
		for {
			select {
			case suc := <-p.Successes():
				fmt.Println("offset: ", suc.Offset, "timestamp: ", suc.Timestamp.String(), "partitions: ", suc.Partition)
			case fail := <-p.Errors():
				log.Println("err: ", fail.Err)
			}
		}
	}(producer)

	// var value string //发送数据
	// for i := 0; ; i++ {
	// 	time.Sleep(10 * time.Millisecond)
	// 	time11 := time.Now()
	// 	value = "这是主题0606_test的一个消息：" + time11.Format("15:04:05")

	// 发送的消息,主题。
	// 注意：这里的msg必须得是新构建的变量，不然你会发现发送过去的消息内容都是一样的，因为批次发送消息的关系。
	// msg := &sarama.ProducerMessage{
	// 	Topic: "0606_test",
	// 	Value: sarama.ByteEncoder(value),
	// 	Key:   sarama.StringEncoder("KEY"),
	// }

	//发送消息
	for {
		message := <-mas
		producer.Input() <- &message
	}

}

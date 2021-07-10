package service

import (
	"github.com/Jundong-chan/seckill/model"
	"github.com/Jundong-chan/seckill/plugins"
	"fmt"
	"sync"
)

var ConsumerWG sync.WaitGroup
var ProducterWG sync.WaitGroup

//开启生产者线程AsyncProducer等待读取通道传来的数据
//开启消费者线程Consumer，读取kafka的数据，再开启异步线程
type KafkaService interface {
	ProduceMessage(topic string, value model.Order)
	AsyncProducer()      //读取channel里的数据，打包成消息发送到kafka
	Consumer()           //从kafka中读取消息，调用订单生成
	InitProductHandle()  //初始化生产者协程
	InitConsumerHandle() //初始化消费者协程
}
type KafkaSvcimpl struct{}

func NewkafkaSvcimpl() *KafkaSvcimpl {
	return &KafkaSvcimpl{}
}

//将数据放入channel
func (ksvc KafkaSvcimpl) ProduceMessage(topic string, value model.Order) {
	kimpl := plugins.NewKafkaProducerimpl()
	kimpl.GenerateMessage(topic, value)
}

//异步生产者线程
func (ksvc KafkaSvcimpl) AsyncProducer() {
	kimpl := plugins.NewKafkaProducerimpl()
	kimpl.AsyncProducer()
}

//消费者线程
func (ksvc KafkaSvcimpl) Consumer() {
	kimpl := plugins.NewKafkaConsumersvcimpl()
	kimpl.CreateOrderConsumer()
}

//异步订单消息生成线程初始化
func (ksvc KafkaSvcimpl) InitProductHandle() {

	for i := 0; i < 5; i++ {
		ProducterWG.Add(1)
		go ksvc.AsyncProducer()
	}
	fmt.Println("生产者开始生产订单消息....")
	ConsumerWG.Wait()
}

//异步订单处理线程初始化
func (ksvc KafkaSvcimpl) InitConsumerHandle() {

	for i := 0; i < 5; i++ {
		ConsumerWG.Add(1)
		go ksvc.Consumer()
	}
	fmt.Println("消费者开始消费订单消息....")
	ConsumerWG.Wait()
}

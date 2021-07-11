package service

import (
	"../model"
	"../plugins"
	"fmt"
	"log"
	"sync"
)

type OrderService interface {
	CreateOrder()
	HandleOrderRequest()
	HandleOrderRequestGoroutine()
}

type OrderServiceImpl struct {
}

func NewOrderSvcimpl() *OrderServiceImpl {
	return &OrderServiceImpl{}
}

func (osvc OrderServiceImpl) CreateOrder(order model.Order) error {
	omodel := model.NewOrderModel()
	err := omodel.CreateOrder(order)
	return err
}

//处理kafka传过来的订单异步消息,存到数据库
func (osvc OrderServiceImpl) HandleOrderRequest() {
	for msg := range plugins.Preordermsg {
		fmt.Printf("正在处理订单消息..... ")
		err := osvc.CreateOrder(msg)
		if err != nil {
			log.Printf("CreateOrder: %v", err)
		}
	}
}

//开启线程
func (osvc OrderServiceImpl) HandleOrderRequestGoroutine() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go osvc.HandleOrderRequest()
	}
	wg.Wait()
	fmt.Println("读取消息的订单正在处理....")
}
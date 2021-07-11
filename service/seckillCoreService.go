package service

//秒杀核心服务
import (
	"github.com/Jundong-chan/seckill/config"
	"github.com/Jundong-chan/seckill/model"
	"errors"
	"fmt"
	"log"
	"strconv"
)

type SeckillCoreService interface {
	ExecuteSeckill(*config.SecRequest) (*config.SecResult, error)
}

type SeckillCoreServiceimpl struct {
}

func (svc SeckillCoreServiceimpl) ExecuteSeckill(req *config.SecRequest) (*config.SecResult, error) {
	conn := model.Pool.Get()
	defer conn.Close()
	//判断商品状态
	status := model.ReadStatus(req.ProductId, conn)
	if status != 1 {
		log.Printf("product :%v ,is off the shell", req.ProductId)
		return &config.SecResult{
			ProductId: req.ProductId,
			UserId:    req.UserId,
			Code:      config.NotSelling,
		}, errors.New("the product is not on selling")
	}
	//判断库存是否可用
	storage := model.ReadStorage(req.ProductId, conn)
	if storage <= 0 || storage < req.BuyNum {
		log.Printf("product :%v ,is off the shell", req.ProductId)
		return &config.SecResult{
			ProductId: req.ProductId,
			UserId:    req.UserId,
			Code:      config.Soldout,
		}, errors.New("the product is not on selling")
	}
	//执行减库存
	err := model.ChangeStorage(req.ProductId, -req.BuyNum, conn)
	if err != nil {
		return &config.SecResult{
			ProductId: req.ProductId,
			UserId:    req.UserId,
			Code:      config.ServiceBusyErr,
		}, err
	}
	//更新用户购买数量的名单
	id := req.ProductId + req.UserId
	actualnum, ok := config.BlackList.Load(id)
	acnum := actualnum.(int)
	//如果没有
	if !ok {
		config.BlackList.Store(id, req.BuyNum)
	} else {
		config.BlackList.Store(id, acnum+req.BuyNum)
	}

	//异步处理订单
	pid, _ := strconv.Atoi(req.ProductId)
	order := model.Order{
		UserId:    req.UserId,
		ProductId: pid,
		Quantity:  req.BuyNum,
	}
	kafkasvc := NewkafkaSvcimpl()
	kafkasvc.ProduceMessage("order", order)

	//购买成功
	fmt.Printf("user: %v buying product: %v success\n", req.UserId, req.ProductId)
	return &config.SecResult{
		ProductId: req.ProductId,
		UserId:    req.UserId,
		Code:      config.Success,
	}, nil
}

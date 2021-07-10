package service
import (
	"CommoditySpike/server/seckillcore/config"
	"CommoditySpike/server/seckillcore/pkg/redis"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
)

//这是接收客户请求的入口处理函数，将请求推入redis，并取出结果

type SeckillService interface {
	SecKill(req *config.SecRequest) (interface{}, error)
	SendSeckillReqEnd() //将请求推入redis
	ReadSeckillResEnd() //将结果从redis读取出来
}
type SeckillServiceimpl struct{}

//接入transport过来的请求，放入channel，等待结果
func (sv SeckillServiceimpl) SecKill(req *config.SecRequest) (interface{}, error) {
	//defer redisclient.UpdateStorageToBase(req.ProductId)

	//查看用户的资格，购买是否到上限
	id := req.ProductId + req.UserId
	fmt.Println("用户请求id", id)
	limit := redisclient.ReadLimit(req.ProductId)
	counts, _ := config.BlackList.Load(id) //查询名单上的购买次数
	var count int
	if counts == nil {
		count = 0
	} else {
		count = counts.(int)
	}
	if count >= limit {
		return config.SecResult{
			ProductId: req.ProductId,
			UserId:    req.UserId,
			Code:      config.Buylimit,
		}, errors.New("you have getting buy limit")
	}
	conn := redisclient.Pool.Get()
	defer conn.Close()
	//查询商品是否在秒杀
	status := redisclient.ReadStatus(req.ProductId, conn)
	if status != 1 {
		return config.SecResult{
			ProductId: req.ProductId,
			UserId:    req.UserId,
			Code:      config.NotSelling,
		}, errors.New("the product is not selling")
	}

	//创建 接收这个用户请求的结果的通道
	resultchan := make(chan *config.SecResult, 1)

	config.SecEndContext.UserConnMapLock.Lock()
	config.SecEndContext.UserConnMap[id] = resultchan
	config.SecEndContext.UserConnMapLock.Unlock()

	defer func() {
		config.SecEndContext.UserConnMapLock.Lock()
		delete(config.SecEndContext.UserConnMap, id)
		config.SecEndContext.UserConnMapLock.Unlock()
	}()

	//将请求送到requestchan
	config.SecEndContext.SecReqChan <- req

	//等待从请求通道接收结果
	ticker := time.NewTicker(time.Duration(8 * time.Second))
	select {
	case <-ticker.C:
		log.Println("endpoint: waiting for result time out")
		return config.SecResult{
			ProductId: req.ProductId,
			UserId:    req.UserId,
			Code:      config.ProccessTimeout,
		}, errors.New("waiting for result time out")
	case result := <-resultchan:
		if result.Code == config.Success {
			fmt.Println("endpoint: a successful buyer")
		}

		return *result, nil
	}

}

//一直循环等待将 端点的请求队列的数据放到redis中
func (sv SeckillServiceimpl) SendSeckillReqEnd() {
	//fmt.Println("writing request into redis.....")
	conn := redisclient.Pool.Get()
	defer conn.Close()
	for {
		req := <-config.SecEndContext.SecReqChan
		data, err := json.Marshal(req)
		if err != nil {
			log.Printf("marshal result failed,%v", err)
			continue
		}
		_, err = conn.Do("LPush", config.RedisHandleReqListName, data)
		if err != nil {
			continue
		}
		fmt.Printf("lpush req success. req : %v", string(data))
	}

}

//将redis队列中的结果取出，放到结果channel
func (sv SeckillServiceimpl) ReadSeckillResEnd() {
	//fmt.Println("endpoint :reading result from redis.....")
	conn := redisclient.Pool.Get()
	defer conn.Close()

	for {
		data, err := redis.Strings(conn.Do("BRPOP", config.RedisHandleResListName, 3))
		if err != nil {
			//log.Println("endpoint :pop a result from redis failed")
			continue
		}
		fmt.Println("从redis结果队列中取出结果：", data)
		//将data反序列化
		var res config.SecResult
		err = json.Unmarshal([]byte(data[1]), &res)
		if err != nil {
			log.Println("endpoint :Unmarshal a result from redis failed")
			continue
		}
		//把data送回指定的channel

		id := res.ProductId + res.UserId
		config.SecEndContext.UserConnMapLock.Lock()
		userchan := config.SecEndContext.UserConnMap[id]
		config.SecEndContext.UserConnMapLock.Unlock()

		fmt.Println("正在等待service 取结果....")
		userchan <- &res

		fmt.Println("endpoint :Send result to userchan success")
	}

}

//开启 将请求推入redis 和 从redis读取结果的协程
func (sv SeckillServiceimpl) InitRedisResAndReq() {
	var sevimpl = SeckillServiceimpl{}

	for i := 0; i < config.PushRequestIntoRedisGoroutineNum; i++ {
		config.RedisHandleWg.Add(1)
		go sevimpl.SendSeckillReqEnd()
	}
	for i := 0; i < config.ReadResultFromRedisGoroutineNum; i++ {
		config.RedisHandleWg.Add(1)
		go sevimpl.ReadSeckillResEnd()
	}
	config.RedisHandleWg.Wait()
	fmt.Println("正在开启处理请求和读取结果的协程....")
}

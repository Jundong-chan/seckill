package service
import (
	"encoding/json"
	"fmt"
	"github.com/Jundong-chan/seckill/config"
	"github.com/Jundong-chan/seckill/model"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
)

type RedisHandleService interface {
	RunProccess()
	HandleReader() //循环将redis请求队列里读取请求，放入处理队列中等待被处理
	HandleUser()   //循环，将处理队列的收到的请求送给秒杀核心，等待秒杀核心处理的结果，并将结果送回结果处理队列
	HandleWrite()  //循环，将结果处理队列中的结果写进redis结果队列中
}

type RedisHandleServiceimpl struct{}

func (rhsrv RedisHandleServiceimpl) RunProccess() {
	for i := 0; i < config.RedisHandleReaderGoroutineNum; i++ {
		config.SecHandleCtx.WaitGroup.Add(1)
		go rhsrv.HandleReader()
	}
	for i := 0; i < config.RedisHandleUserGoroutineNum; i++ {
		config.SecHandleCtx.WaitGroup.Add(1)
		go rhsrv.HandleUser()
	}
	for i := 0; i < config.RedisHandleWriterGoroutineNum; i++ {
		config.SecHandleCtx.WaitGroup.Add(1)
		go rhsrv.HandleWrite()
	}

	fmt.Println("所有的处理协程已经开启")
	config.SecHandleCtx.WaitGroup.Wait()

}

//读取redis 队列中的请求，推给秒杀接收请求的队列
func (rhsrv RedisHandleServiceimpl) HandleReader() {
	fmt.Println("正在读取redis中的请求...")
	rdb := model.Pool.Get()
	defer rdb.Close()
	for {
		for {
			//获取redis队列里的请求
			data, err := redis.Strings(rdb.Do("BRPOP", config.RedisHandleReqListName, 3))
			if err != nil {
				//log.Println("HandleReader brpop failed: " + err.Error())
				continue
			}
			fmt.Println("brpop from request list: ", data[1])

			//转换数据结构
			var req config.SecRequest
			err = json.Unmarshal([]byte(data[1]), &req)
			if err != nil {
				log.Println("Unmarshal request data failed: ", err.Error())
			}
			//判断这个请求是否超时
			if (time.Now().Unix() - req.SecTime) >= int64(config.MaxRequestWaitTimeout) {
				log.Printf("the request req[%v]is expired\n", req)
				continue
			}
			//设置等待结果的超时时间
			timer := time.NewTicker(time.Millisecond * time.Duration(config.MaxWaitResultTimeout))
			select {
			case config.SecHandleCtx.ReadHandleChan <- &req:
			case <-timer.C:
				log.Printf("send to handle chan failed , req[%v]", req)
				// break
			}
		}
	}
}

//中间人，接收redis推过来的请求，送到秒杀核心处理
func (rhsrv RedisHandleServiceimpl) HandleUser() {
	fmt.Println("请求处理中.....")
	seckillsvc := SeckillCoreServiceimpl{}

	for req := range config.SecHandleCtx.ReadHandleChan {
		result, err := seckillsvc.ExecuteSeckill(req)
		if err != nil {
			log.Printf("Excute request failed: %v", err)
		}
		timer := time.NewTicker(time.Millisecond * time.Duration(config.SendToWriteChanTimeout))
		select {
		case config.SecHandleCtx.WriteHandleChan <- result:
		case <-timer.C:
			log.Printf("Send to resukt channel timeout, res: %v", result)
			// break
		}
	}
}

//将结果放回redis
func (rhsrv RedisHandleServiceimpl) HandleWrite() {
	fmt.Println("sending result to redislist")
	for res := range config.SecHandleCtx.WriteHandleChan {
		err := SendToRedis(res)
		if err != nil {
			log.Printf("send result to redis failed: %v", err)
			continue
		}
	}
}

func SendToRedis(res *config.SecResult) error {
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("marshal result failed,%v", err)
		return err
	}

	rdb := model.Pool.Get()
	defer rdb.Close()
	_, err = rdb.Do("LPush", config.RedisHandleResListName, data)
	if err != nil {
		log.Printf("failed to push result to redis res: %v", data)
		return err
	}
	fmt.Printf("push data into redislist success\n")
	return nil
}

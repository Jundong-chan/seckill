package main

import (
	"./endpoint"
	"./pkg"
	"./service"
	"./transport"
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"./model"
)

func main() {
	var (
		//参数说明：命令行参数名称，默认值，参数说明（如果使用-help会给出这个说明提示）
		//使用变量时要加 *
		serverhost    = flag.String("host", "", "server host")
		serverport    = flag.String("port", "8810", "server port")
		mysqluser     = flag.String("mysqluser", "root", "mysql user")
		mysqlpassword = flag.String("mysqlpassword", "gzhuchan", "mysql password")
		mysqlhost     = flag.String("mysqlhost", "127.0.0.1", "mysql host")
		mysqlport     = flag.String("mysqlport", "3306", "mysql port")
		mysqlschema   = flag.String("mysqlschema", "Seckill", "mysql schema")
		RedisAddr     = flag.String("redisaddr", "127.0.0.1:6379", "redis connect address")
	)
	//初始化
	flag.Parse() //解析命令行参数
	pkg.Init(*mysqluser, *mysqlpassword, *mysqlhost, *mysqlport, *mysqlschema)
	model.RedisInit(*RedisAddr) //连接redis和更新mysql数据到redis

	var redishandle = service.RedisHandleServiceimpl{}
	go func() {
		redishandle.RunProccess() //启动redis 处理请求和结果的list
	}()

	ctx := context.Background()

	var serv1 = service.SeckillServiceimpl{}
	var serv2 = service.CheckServiceimpl{}

	//启动处理redis 队列的函数
	go serv1.InitRedisResAndReq()

	//开启异步处理订单的所有线程
	kafkasvc := service.NewkafkaSvcimpl()
	osvc := service.NewOrderSvcimpl()
	go kafkasvc.AsyncProducer()
	go kafkasvc.Consumer()
	go osvc.HandleOrderRequestGoroutine()

	var SeckillServEp = endpoint.MakeSeckillServiceEp(serv1)
	var HealthServEp = endpoint.MakeHealthCheckEp(serv2)

	var SecEp = endpoint.SeckillServiceEndpoint{
		SeckillServiceEp: SeckillServEp,
		HealthCheckEp:    HealthServEp,
	}
	h := transport.MakeHttpHandler(ctx, SecEp)
	errchan := make(chan error)
	go func() {
		fmt.Println("Http Server start at port:" + (*serverport))
		errchan <- http.ListenAndServe(*serverhost+":"+*serverport, h)
	}()
	fmt.Println("waiting")
	<-errchan
}


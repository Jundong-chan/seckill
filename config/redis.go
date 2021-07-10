package config

import (
	"github.com/Jundong-chan/seckill/model"
	"sync"
)

var RedisHandleReqListName = "RequestList"

var RedisHandleResListName = "ResultList"

var RedisHandleReaderGoroutineNum = 10

var RedisHandleWriterGoroutineNum = 10

var RedisHandleUserGoroutineNum = 10

var PushRequestIntoRedisGoroutineNum = 300

var ReadResultFromRedisGoroutineNum = 300

var RedisHandleWg sync.WaitGroup

//由商品id与商品信息构成的map,存放从redis中读取的商品数据
var Secproductinfo = make(map[string]model.Product, 4096)
var ProductInfoMapRWLocker sync.Mutex
var BlackList sync.Map //黑名单,暂时先将这个名单放在内存中, key是 pid+userid

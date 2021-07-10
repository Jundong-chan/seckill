package config

import (
	"sync"
	"time"
)

var MaxRequestWaitTimeout = 5 * time.Second //设定请求超时的时间 s
var MaxWaitResultTimeout = 600              //等待放入处理队列的时间 ms
var SendToWriteChanTimeout = 600            //等待协程将结果放入结果队列的超时时间 ms

// //var SecProductInfoMap SecProductInfo

type SecRequest struct {
	ProductId   string          `json:"productid"` //商品ID
	SecTime     int64           `json:"sectime"`   //用户点击的时间
	UserId      string          `json:"userid"`
	AccessTime  int64           `json:"accesstime"`
	BuyNum      int             `json:"buynum"` //用户请求购买的数量
	Price       float64         `json:"price"`
	ClientAddr  string          `json:"clientaddr"` //用户的请求ip
	CloseNotify <-chan bool     `json:"-"`          //用来获知用户提前关闭请求
	ResultChan  chan *SecResult `json:"-"`
}

type SecResult struct {
	ProductId string `json:"productid"` //商品ID
	UserId    string `json:"userid"`    //用户ID
	Code      int    `json:"code"`      //结果的状态码
}

const (
	Success = iota
	NotSelling
	ServiceBusyErr
	Soldout
	Buylimit
	ProccessTimeout
)

//秒杀接收层的上下文内容
type SecEndpointContext struct {
	SecReqChan       chan *SecRequest
	RWSecProductLock sync.RWMutex

	UserConnMap     map[string]chan *SecResult //存放用户接收结果的chan
	UserConnMapLock sync.Mutex
}

var SecEndContext = SecEndpointContext{
	SecReqChan:  make(chan *SecRequest, 4096),
	UserConnMap: make(map[string]chan *SecResult, 4096),
}

//秒杀redis处理层的上下文内容
type SecHandleContext struct {
	ReadHandleChan  chan *SecRequest
	WriteHandleChan chan *SecResult
	RwProductStor   sync.RWMutex
	WaitGroup       sync.WaitGroup
	//ProductCountMag *model.ProductStorageManage //商品计数存储
}

//初始化变量
var SecHandleCtx = &SecHandleContext{
	ReadHandleChan:  make(chan *SecRequest, 10000),
	WriteHandleChan: make(chan *SecResult, 10000),
	//ProductCountMag: model.NewProductCountMgr(),
}

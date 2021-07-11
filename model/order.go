package model

import (
	"../pkg"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/gohouse/gorose"
)

var Serverid = "1" //服务器id

type Order struct {
	OrderId     int64   `gorose:"orderid" json:"orderid"`
	Ordersn     string  `gorose:"order_sn" json:"ordersn"`
	UserId      string  `gorose:"user_id" json:"userid"`
	ProductId   int     `gorose:"product_id" json:"product_id"`
	Status      int     `gorose:"status" json:"status"`
	Quantity    int     `gorose:"quantity" json:"quantity"`
	TotalAmount float64 `gorose:"total_amount" json:"total_amount,string"`
	CreateTime  string  `gorose:"create_time" json:"create_time"`
	UpdateTime  string  `gorose:"update_time" json:"update_time"`
}

type Ordermodel struct{}

func NewOrderModel() *Ordermodel {
	return &Ordermodel{}
}

func (svc *Ordermodel) getTableName() string {
	return "orders"
}

func (svc *Ordermodel) CreateOrder(order Order) error {
	conn := pkg.DB()
	s1 := rand.NewSource(time.Now().Unix())
	s2 := rand.New(s1)
	ran := strconv.Itoa(s2.Intn(1000) + 1000) //产生一个1000~2000的随机数
	uid := order.UserId[len(order.UserId)-4:]
	pid := strconv.Itoa(order.ProductId)
	tnano := strconv.FormatInt(time.Now().UnixNano(), 10)
	nano := tnano[len(tnano)-5:]
	//最终生成的订单格式 当前时间+当前纳秒最后5位+用户最后4位id+商品id+一个随机数+服务器id
	order.Ordersn = time.Now().Format("20060102150405") + nano + uid + pid + ran + Serverid
	order.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	order.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	//默认订单状态是已经支付的
	if order.Status == 0 {
		order.Status = 1
	}
	pmodel := NewProductModel()
	product, _ := pmodel.QureyProductByUserId("product_id", "=", order.ProductId)

	//计算总价格
	price, _ := strconv.ParseFloat(product[0]["price"].(string), 64)
	order.TotalAmount = price * float64(order.Quantity)
	//保留两位小数
	order.TotalAmount = math.Trunc(order.TotalAmount*1e2+0.5) * 1e-2
	fmt.Println("price is ,TotalAmount is : ", price, order.TotalAmount)

	_, err := conn.Table(svc.getTableName()).Data(order).Insert()
	if err != nil {
		return errors.New("create order failed :" + err.Error())
	}
	fmt.Printf("创建订单成功，编号: %v 订单用户: %v ,订单商品: %v , 订单数量: %v, 订单总额: %v",
		order.Ordersn,
		order.UserId,
		order.ProductId,
		order.Quantity,
		order.TotalAmount,
	)
	return nil
}

//根据条件查询订单
func (svc *Ordermodel) QureyOrderByConndition(field string, condition string, value interface{}) ([]gorose.Data, error) {

	if condition == "" {
		condition = ""
	}
	conn := pkg.DB()
	data, err := conn.Table(svc.getTableName()).Where(field, condition, value).Get()
	if err != nil {
		return nil, errors.New("qurey order failed :" + err.Error())
	}
	return data, nil
}

//将需要修改的字段和值放进map , 根据查询条件 field=value时
func (svc *Ordermodel) ModifyOrderByfield(data map[string]interface{}, field string, value interface{}) error {
	conn := pkg.DB()
	data["update_time"] = time.Now().Format("2006-01-02 15:04:05")
	_, err := conn.Table(svc.getTableName()).Data(data).Where(field, value).Update()
	if err != nil {
		return errors.New("ModifyOrderByfield failed :" + err.Error())
	}
	return nil
}

func (svc *Ordermodel) QureyOrderlist() ([]gorose.Data, error) {
	conn := pkg.DB()
	data, err := conn.Table(svc.getTableName()).Get()
	if err != nil {
		return nil, errors.New("QureyOrderlist failed :" + err.Error())
	}
	return data, nil
}

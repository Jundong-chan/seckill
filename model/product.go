package model

import (
	"../pkg"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gohouse/gorose"
)

//这是product数据库映射

type Product struct {
	ProductId    int     `gorose:"product_id" json:"product_id"` //数据库自增
	ProductName  string  `gorose:"product_name" json:"product_name"`
	Price        float64 `gorose:"price" json:"price,string"`
	Total        int     `gorose:"total" json:"total"`
	Status       int     `gorose:"status" json:"status"`
	Detail       string  `gorose:"detail" json:"detail"`
	ProductOwner string  `gorose:"product_owner" json:"product_owner"`
	CreateTime   string  `gorose:"create_time" json:"create_time"`
	UpdateTime   string  `gorose:"update_time" json:"update_time"`
	SecstartTime string  `gorose:"secstart_time" json:"secstart_time"`
	SecendTime   string  `gorose:"secend_time" json:"secend_time"`
	Buylimit     int     `gorose:"buylimit" json:"buylimit"`
}

type Productmodel struct {
}

func NewProductModel() *Productmodel {
	return &Productmodel{}
}

func (pmodel *Productmodel) getTableName() string {
	return "product"
}

func (pmodel *Productmodel) GetProductList() ([]gorose.Data, error) {
	list, err := pkg.DB().Table(pmodel.getTableName()).Get()
	if err != nil {
		log.Printf("GetProductList error:%v", err)
		return nil, err
	}
	return list, nil
}

func (pmodel *Productmodel) GetSecProductInfo() ([]gorose.Data, error) {
	conn := pkg.DB()
	list, err := conn.Table(pmodel.getTableName()).Where("status", 1).Get()
	if err != nil {
		log.Printf("GetSecProductInfo error:%v", err)
		return nil, err
	}
	return list, nil
}

func (pmodel *Productmodel) CreateProduct(p *Product) error {
	conn := pkg.DB()
	p.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	p.UpdateTime = time.Now().Format("2006-01-02 15:04:05")

	//将前端设置的商品秒杀时间戳转成 日期格式
	secstart, err1 := strconv.Atoi(p.SecstartTime)
	secend, err2 := strconv.Atoi(p.SecendTime)
	if err1 != nil || err2 != nil {
		return errors.New("time format wrong")
	}
	secstart_time := time.Unix(int64(secstart), 0).Format("2006-01-02 15:04:05")
	secend_time := time.Unix(int64(secend), 0).Format("2006-01-02 15:04:05")

	_, err := conn.Table(pmodel.getTableName()).Data(
		map[string]interface{}{
			"product_name":  p.ProductName,
			"price":         p.Price,
			"total":         p.Total,
			"status":        p.Status,
			"detail":        p.Detail,
			"product_owner": p.ProductOwner,
			"create_time":   p.CreateTime,
			"update_time":   p.UpdateTime,
			"secstart_time": secstart_time,
			"secend_time":   secend_time,
			"buylimit":      p.Buylimit,
		},
	).Insert()
	if err != nil {
		return err
	}
	//创建商品后需要更新到redis
	//redisclient.ReadDataFromMysql()
	return nil
}

//删除 商品
func (pmodel *Productmodel) DeleteProduct(id string) error {
	conn := pkg.DB()
	ID, _ := strconv.Atoi(id)
	_, err := conn.Table(pmodel.getTableName()).Where("product_id", ID).Delete()
	if err != nil {
		return err
	}
	return nil
}

//修改商品的个别信息
func (pmodel *Productmodel) ModifyProductByCondition(data map[string]interface{}, condition string, mes interface{}) error {
	conn := pkg.DB()
	UpdateTime := time.Now().Format("2006-01-02 15:04:05")
	data["update_time"] = UpdateTime
	fmt.Println("update data is :", data)
	_, err := conn.Table(pmodel.getTableName()).Data(data).Where(condition, mes).Update()
	if err != nil {
		log.Println("update product strage to mysql failed")
		return err
	}
	return nil
}

//修改,秒杀商品的信息
func (pmodel *Productmodel) ModifyProduct(p *Product) error {
	conn := pkg.DB()
	p.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	//将前端设置的商品秒杀时间戳转成 日期格式
	secstart, err1 := strconv.Atoi(p.SecstartTime)
	secend, err2 := strconv.Atoi(p.SecendTime)
	if err1 != nil || err2 != nil {
		return errors.New("time format wrong")
	}
	secstart_time := time.Unix(int64(secstart), 0).Format("2006-01-02 15:04:05")
	secend_time := time.Unix(int64(secend), 0).Format("2006-01-02 15:04:05")

	_, err := conn.Table(pmodel.getTableName()).Data(
		map[string]interface{}{
			"product_name":  p.ProductName,
			"price":         p.Price,
			"total":         p.Total,
			"status":        p.Status,
			"detail":        p.Detail,
			"update_time":   p.UpdateTime,
			"secstart_time": secstart_time,
			"secend_time":   secend_time,
			"buylimit":      p.Buylimit,
		},
	).Where("product_id", p.ProductId).Update()
	if err != nil {
		return err
	}
	return nil
}

func (pmodel *Productmodel) QureyProductByUserId(feild string, condition string, value interface{}) ([]gorose.Data, error) {
	if condition == "" {
		condition = "="
	}
	conn := pkg.DB()
	fmt.Println("feild:", feild, "condition", condition, "value", value)
	result, err := conn.Table(pmodel.getTableName()).Where(feild, condition, value).Get()
	if err != nil {
		log.Printf("QureyProductByUserId Error:%v", err)
		return nil, err
	}
	return result, nil
}

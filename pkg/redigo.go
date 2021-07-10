package pkg

import (
    "github.com/Jundong-chan/seckill/model"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

//redigo数据库
// var RedisAddr = "127.0.0.1:6379"
var Pool *redis.Pool
var RsProductstorekey = "productstore"
var RsProductstatuskey = "productstatus"
var RsProductstartkey = "productstart"
var RsProductendkey = "productend"
var RsProductlimitkey = "buylimit"
var StorageLock sync.RWMutex

func newRedisPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     5000,             //最大空闲连接数
		MaxActive:   7000,             //最大连接数
		IdleTimeout: 10 * time.Second, //等待超时的时间
		// Dial or DialContext must be set. When both are set, DialContext takes precedence over Dial.
		Dial: func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}

}

func RedisInit(addr string) {
	Pool = newRedisPool(addr)
	ReadDataFromMysql()
}

//将所有商品的库存，商品状态，开始时间，结束时间，读取到redis中
func ReadDataFromMysql() {
	pm := model.NewProductModel()
	data, err := pm.GetProductList() //读取商品信息
	if err != nil {
		log.Println("ReadDataFromMysql failed")
		return
	}
	fmt.Println("读取的data:", data)
	var (
		id      int64
		sid     string
		storage int64
		status  int64
		start   string
		end     string
		limit   int64
	)
	conn := Pool.Get()
	defer conn.Close()

	fmt.Println("reading productinfo to redis....")
	for k := range data {
		//只读取在售的商品
		sta := data[k]["status"].(int64)
		if sta != 1 {
			continue
		}

		id = data[k]["product_id"].(int64)
		sid = strconv.FormatInt(id, 10)

		storage = data[k]["total"].(int64)
		_, err := conn.Do("HSet", RsProductstorekey, sid, storage)
		if err != nil {
			log.Println("setting data to redis failed")
		}
		status = data[k]["status"].(int64)
		_, err = conn.Do("HSet", RsProductstatuskey, sid, status)
		if err != nil {
			log.Println("setting data to redis failed")
		}

		start = data[k]["secstart_time"].(string)
		_, err = conn.Do("HSet", RsProductstartkey, sid, start)
		if err != nil {
			log.Println("setting data to redis failed")
		}

		end = data[k]["secend_time"].(string)
		_, err = conn.Do("HSet", RsProductendkey, sid, end)
		if err != nil {
			log.Println("setting data to redis failed")
		}
		limit = data[k]["buylimit"].(int64)
		fmt.Println("limit : ", limit)
		_, err = conn.Do("HSet", RsProductlimitkey, sid, limit)
		if err != nil {
			log.Println("setting data to redis failed")
		}
	}
	log.Println("success write data to redis ")
}

//将redis商品的库存信息更新到mysql
func UpdateStorageToBase(pid string) {
	conn := Pool.Get()
	defer conn.Close()
	//fmt.Println("writing productinfo to mysql")

	stor, err := redis.Int(conn.Do("HGet", RsProductstorekey, pid))
	if err != nil {
		log.Println("setting data to redis failed")
	}
	fmt.Println("修改后剩下的库存：", stor)
	status, err := redis.Int(conn.Do("HGet", RsProductstatuskey, pid))
	if err != nil {
		log.Println("setting data to redis failed")
	}
	//fmt.Println("修改后的状态：", status)

	data := map[string]interface{}{
		"total":  stor,
		"status": status,
	}
	pm := model.NewProductModel()
	err = pm.ModifyProductByCondition(data, "product_id", pid)
	if err != nil {
		log.Println("failed to write data to mysql ")
	}
	log.Println("success write data to mysql ")
}

//读取库存
func ReadStorage(productid string, conn redis.Conn) int {
	// conn := Pool.Get()
	// defer conn.Close()
	data, err := redis.Int(conn.Do("HGet", RsProductstorekey, productid))
	//log.Printf("reading product storage :%v", data)
	if err != nil {
		fmt.Println("get storage from redis failed")
		return 0
	}
	return data

}

//读取状态
func ReadStatus(productid string, conn redis.Conn) int {
	// conn := Pool.Get()
	// defer conn.Close()
	data, err := redis.Int(conn.Do("HGet", RsProductstatuskey, productid))
	//log.Printf("reading productstatus :%v", data)
	if err != nil {
		fmt.Println("get product_status from redis failed")
		return 0
	}
	return data
}

//读取秒杀开始时间
func ReadStart(productid string) int64 {
	conn := Pool.Get()
	defer conn.Close()
	data, err := redis.String(conn.Do("HGet", RsProductstartkey, productid))
	//log.Printf("reading secstarttime :%v", data)
	if err != nil {
		fmt.Println("get starttime from redis failed")
		return 0
	}
	t, _ := time.Parse("2006-01-02 15:05:04", data)
	return t.Unix()
}

//读取秒杀结束时间
func ReadEnd(productid string) int64 {
	conn := Pool.Get()
	defer conn.Close()
	data, err := redis.String(conn.Do("HGet", RsProductendkey, productid))
	//log.Printf("reading secendtime :%v", data)
	if err != nil {
		fmt.Println("get endtime from redis failed")
		return 0
	}
	t, _ := time.Parse("2006-01-02 15:05:04", data)
	return t.Unix()
}

//读取商品的购买数量限制
func ReadLimit(productid string) int {
	conn := Pool.Get()
	defer conn.Close()
	data, err := redis.Int(conn.Do("HGet", RsProductlimitkey, productid))
	//log.Printf("reading productlimit :%v", data)
	if err != nil {
		fmt.Println("get product limit failed")
		return 0
	}
	return data
}

//修改库存
func ChangeStorage(productid string, num int, conn redis.Conn) error {
	// conn := Pool.Get()
	// defer conn.Close()
	//再次查询库存看是否够
	store := ReadStatus(productid, conn)
	if store-num <= 0 {
		log.Printf("product: %v sold out", productid)
		return errors.New("sold out")
	}
	nums := int64(num)
	result, err := conn.Do("HIncrBy", RsProductstorekey, productid, nums)
	if err != nil {
		log.Println("decrese storage failed")
		return errors.New("decrese storage failed ")
	}
	fmt.Println("修改后剩下的库存：", result)
	return nil
}

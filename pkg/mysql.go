package pkg

//这是mysql初始化配置包，使用gorose封装了mysql具体操作
import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose"
)

var err error
var engin *gorose.Engin

func Init(user, password, host, port, schema string) {
	//Dsn格式：数据库用户名+:+密码+tcp(@服务器ip:数据库端口号)/schema名？+配置参数
	engin, err = gorose.Open(&gorose.Config{
		Driver:          "mysql",
		Dsn:             user + ":" + password + "@tcp(" + host + ":" + port + ")/" + schema + "?charset=utf8mb4&&parseTime=true",
		SetMaxOpenConns: 500, //数据库最大连接数，默认0表示无限制
		SetMaxIdleConns: 10,  //数据库最大空闲连接数，默认为1
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	//数据库只需要被open一次，不需要担心数据库连接被耗尽，go提供的数据库驱动会维护数据库连接池
}

//常用于返回一个新的 ORM对象来操控配置好的数据库engin,调用engin.NewOrm()即可操作我们配置好的数据库
func DB() gorose.IOrm {
	return engin.NewOrm()
}

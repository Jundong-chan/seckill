package model

import (
	"../pkg"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gohouse/gorose"
	"github.com/rs/xid"
)

type User struct {
	UserId     string `gorose:"user_id" json:"userid"`
	UserName   string `gorose:"user_name" json:"username"`
	Password   string `gorose:"password"`
	Salt       string `gorose:"salt"`
	Gender     string `gorose:"gender"`
	Email      string `gorose:"email"`
	Phone      string `gorose:"phone"`
	CreateTime string `gorose:"create_time"`
	UpdateTime string `gorose:"update_time"`
}
type Usermodel struct {
}

func NewUserModel() *Usermodel {
	return &Usermodel{}
}

func (umodel *Usermodel) getTableName() string {
	return "seckill_user"
}

func (umodel *Usermodel) GetUserList() ([]gorose.Data, error) {
	conn := pkg.DB()
	list, err := conn.Table(umodel.getTableName()).Get()
	if err != nil {
		log.Printf("GetUserList Error:%v", err)
		return nil, err
	}
	return list, nil
}

func (umodel *Usermodel) CreateUser(user *User) (err error) {
	conn := pkg.DB()
	//生成全局唯一ID
	user.UserId = xid.New().String()
	user.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	user.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(user)

	//随机生成5个字节的盐，盐的位数是 字节数*2
	salt, err := pkg.NewRandomSalt(5)
	user.Salt = salt
	//返回加密后的结果
	user.Password = pkg.Encryption(user.Password, salt)

	if err != nil {
		return err
	}
	//fmt.Println(user.Password, user.Salt)
	_, err = conn.Table(umodel.getTableName()).Data(*user).Insert()
	if err != nil {
		log.Printf("CreateUser Error:%v", err)
		return err
	}
	return nil
}

//使用where查询
func (umodel *Usermodel) QureyUserByCondition(field string, condition string, value interface{}) ([]gorose.Data, error) {
	if condition == "" {
		condition = "="
	}
	conn := pkg.DB()
	result, err := conn.Table(umodel.getTableName()).Where(field, condition, value).Get()
	if err != nil {
		log.Printf("QureyUserByCondition Error:%v", err)
		return nil, err
	}
	return result, nil

}

//传入 用户登陆的字段，用户登陆使用的用户名以及密码，验证用户，正确则返回用户的信息
func (umodel *Usermodel) CheckUser(field string, username string, password string) ([]gorose.Data, error) {
	conn := pkg.DB()
	var user []gorose.Data
	fmt.Println("field", field)
	fmt.Println("username", username)
	fmt.Println("password", password)
	user, err := conn.Table(umodel.getTableName()).Where(field, "=", username).Get()
	if err != nil {
		return nil, errors.New("Wrong field," + err.Error())
	}
	//fmt.Println("qurey user:", user)
	if len(user) == 0 {
		return nil, errors.New("wrong User Account")
	}
	iscorrect := pkg.TestifyEncrypt(password, user[0]["salt"], user[0]["password"])
	if !iscorrect {
		return nil, errors.New("wrong password")
	}
	return user, nil
}

//判断用户是否存在
func (umodel *Usermodel) CheckUserId(userid string) error {
	conn := pkg.DB()
	user, err := conn.Table(umodel.getTableName()).Where("user_id", userid).Get()
	if err != nil {
		return errors.New("wrong User Account")
	}
	if user == nil {
		return errors.New("user doesn't exist")
	}
	return nil

}

package pkg

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

//加密解密密码，验证密码

func NewRandomSalt(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		fmt.Printf("Gnerate Salt error:%v\n", err)
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

//传入明文和盐 使用sha1加密，返回加密后的信息
func Encryption(message string, salt string) (encriped string) {
	sha1hash1 := sha1.New()
	//先对信息加密
	sha1hash1.Write([]byte(message))
	encmeesage := hex.EncodeToString(sha1hash1.Sum(nil))

	//将加密后的信息和盐合并
	mixedmessage := encmeesage + salt
	sha1hash2 := sha1.New()
	sha1hash2.Write([]byte(mixedmessage))
	//对合并后的信息加密
	ecryped := hex.EncodeToString(sha1hash2.Sum(nil))
	return ecryped
}

//传入明文，盐，密文，验证信息是否正确
func TestifyEncrypt(Plaintext string, salt interface{}, ciphertext interface{}) bool {
	salts := salt.(string)
	sha1hash1 := sha1.New()
	sha1hash1.Write([]byte(Plaintext))
	encmeesage := hex.EncodeToString(sha1hash1.Sum(nil))

	mixedmessage := encmeesage + salts
	sha1hash2 := sha1.New()
	sha1hash2.Write([]byte(mixedmessage))
	ecryped := hex.EncodeToString(sha1hash2.Sum(nil))
	if ecryped != ciphertext {
		return false
	}
	return true
}

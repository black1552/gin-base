package utils

import (
	"fmt"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("加密失败:%v", err)
	}
	return string(hashPassword), nil
}

func ValidPassword(hashPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("密码不匹配:%v", err)
	}
	return nil
}

func Md5Password(password string) string {
	md5pass, err := gmd5.EncryptString(password)
	if err != nil {
		return ""
	}
	return md5pass
}

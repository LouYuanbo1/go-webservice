package bcrypt

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// 哈希加密工具,密码禁止明文传递,cost为加密成本，默认10(bcrypt.DefaultCost)
func Encrypt(secret string, cost int) ([]byte, error) {
	// 密码加密
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(secret), cost)
	if err != nil {
		return nil, fmt.Errorf("Encrypt(crypto): 加密失败:%v", err)
	}
	return hashedSecret, nil
}

// 密码校验
func CheckSecret(secret string, hashedSecret []byte) error {
	// 密码校验
	err := bcrypt.CompareHashAndPassword(hashedSecret, []byte(secret))
	if err != nil {
		return fmt.Errorf("CheckSecret(crypto): 校验失败:%v", err)
	}
	return nil
}

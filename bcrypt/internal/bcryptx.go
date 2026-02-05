package internal

import (
	"fmt"

	"github.com/LouYuanbo1/go-webservice/bcrypt/config"
	"golang.org/x/crypto/bcrypt"
)

type bcryptX struct {
	cost int
}

func NewBcryptX(bcryptConfig config.BcryptConfig) *bcryptX {
	return &bcryptX{
		cost: bcryptConfig.Cost,
	}
}

func (b *bcryptX) Encrypt(secret string) ([]byte, error) {
	// 密码加密
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(secret), b.cost)
	if err != nil {
		return nil, fmt.Errorf("Encrypt(crypto): 加密失败:%v", err)
	}
	return hashedSecret, nil
}

func (b *bcryptX) CheckSecret(secret string, hashedSecret []byte) error {
	// 密码校验
	err := bcrypt.CompareHashAndPassword(hashedSecret, []byte(secret))
	if err != nil {
		return fmt.Errorf("CheckSecret(crypto): 校验失败:%v", err)
	}
	return nil
}

package internal

import (
	"fmt"

	"github.com/LouYuanbo1/go-webservice/cryptutil/config"
	"github.com/LouYuanbo1/go-webservice/cryptutil/options"
	"golang.org/x/crypto/bcrypt"
)

type cryptUtil struct {
	defaultCost int
}

func NewCryptUtil(bcryptConfig config.CryptUtilConfig) *cryptUtil {
	return &cryptUtil{
		defaultCost: bcryptConfig.DefaultCost,
	}
}

func (c *cryptUtil) costBuilder(opts ...options.CostOption) int {
	cost := options.Cost{
		Value: c.defaultCost,
	}
	for _, opt := range opts {
		opt(&cost)
	}
	return cost.Value
}

func (c *cryptUtil) Encrypt(secret string, opts ...options.CostOption) ([]byte, error) {
	// 密码加密
	// 密码加密
	cost := c.costBuilder(opts...)
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(secret), cost)
	if err != nil {
		return nil, fmt.Errorf("Encrypt(crypto): 加密失败:%v", err)
	}
	return hashedSecret, nil
}

func (c *cryptUtil) CheckSecret(secret string, hashedSecret []byte) error {
	// 密码校验
	err := bcrypt.CompareHashAndPassword(hashedSecret, []byte(secret))
	if err != nil {
		return fmt.Errorf("CheckSecret(crypto): 校验失败:%v", err)
	}
	return nil
}

package internal

import (
	"github.com/LouYuanbo1/go-webservice/cryptutil/config"
	"github.com/LouYuanbo1/go-webservice/cryptutil/errors"
	"github.com/LouYuanbo1/go-webservice/cryptutil/options"
	"golang.org/x/crypto/bcrypt"
)

type cryptUtil struct {
	defaultCost int
}

func NewCryptUtil(bcryptConfig config.CryptUtilConfig) *cryptUtil {
	return &cryptUtil{
		defaultCost: bcryptConfig.Cost,
	}
}

func (c *cryptUtil) Encrypt(secret string, opts ...options.CostOption) ([]byte, error) {
	// 密码加密
	// 密码加密
	cost := c.costBuilder(opts...)
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(secret), cost.GetCost())
	if err != nil {
		return nil, errors.New(
			errors.ErrEncrypt,
			"Encrypt",
			err,
		)
	}
	return hashedSecret, nil
}

func (c *cryptUtil) CheckSecret(secret string, hashedSecret []byte) error {
	// 密码校验
	err := bcrypt.CompareHashAndPassword(hashedSecret, []byte(secret))
	if err != nil {
		return errors.New(
			errors.ErrEncrypt,
			"CheckSecret",
			err,
		)
	}
	return nil
}

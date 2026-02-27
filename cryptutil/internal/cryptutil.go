package internal

import (
	"github.com/LouYuanbo1/go-webservice/cryptutil/config"
	"github.com/LouYuanbo1/go-webservice/cryptutil/errors"
	"github.com/LouYuanbo1/go-webservice/cryptutil/options"
	"golang.org/x/crypto/bcrypt"
)

type cryptUtil struct {
	config config.CryptUtilConfig
}

func NewCryptUtil(config config.CryptUtilConfig) *cryptUtil {
	return &cryptUtil{
		config: config,
	}
}

func (c *cryptUtil) Encrypt(secret string, opts ...options.CostOption) ([]byte, error) {
	// 密码加密
	cost := options.CostBuilder(c.config, opts...)

	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(secret), cost.GetValue())
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

package cryptutil

import (
	"github.com/LouYuanbo1/go-webservice/errorx"
	"golang.org/x/crypto/bcrypt"
)

type CryptUtil interface {
	Encrypt(secret string, opts ...CostOption) ([]byte, error)
	CheckSecret(secret string, hashedSecret []byte) error
}

type cryptUtil struct {
	config Config
}

func NewCryptUtil(config Config) *cryptUtil {
	return &cryptUtil{
		config: config,
	}
}

func (c *cryptUtil) Encrypt(secret string, opts ...CostOption) ([]byte, error) {
	// 密码加密
	cost := c.costBuilder(opts...)

	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(secret), cost.value)
	if err != nil {
		return nil, errorx.New(
			ErrEncrypt,
			"cryptutil",
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
		return errorx.New(
			ErrCheckSecret,
			"cryptutil",
			"CheckSecret",
			err,
		)
	}
	return nil
}

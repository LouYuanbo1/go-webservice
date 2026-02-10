package cryptutil

import (
	"github.com/LouYuanbo1/go-webservice/cryptutil/config"
	"github.com/LouYuanbo1/go-webservice/cryptutil/internal"
	"github.com/LouYuanbo1/go-webservice/cryptutil/options"
)

type CryptUtil interface {
	Encrypt(secret string, opts ...options.CostOption) ([]byte, error)
	CheckSecret(secret string, hashedSecret []byte) error
}

func NewCryptUtil(config config.CryptUtilConfig) CryptUtil {
	return internal.NewCryptUtil(config)
}

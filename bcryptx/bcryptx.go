package bcryptx

import (
	"github.com/LouYuanbo1/go-webservice/bcryptx/config"
	"github.com/LouYuanbo1/go-webservice/bcryptx/internal"
)

type BcryptX interface {
	Encrypt(secret string) ([]byte, error)
	CheckSecret(secret string, hashedSecret []byte) error
}

func NewBcryptX(bcryptConfig config.BcryptConfig) BcryptX {
	return internal.NewBcryptX(bcryptConfig)
}

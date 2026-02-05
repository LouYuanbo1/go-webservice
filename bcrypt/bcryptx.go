package bcryptx

import (
	"github.com/LouYuanbo1/go-webservice/bcrypt/config"
	"github.com/LouYuanbo1/go-webservice/bcrypt/internal"
)

type BcryptX interface {
	Encrypt(secret string) ([]byte, error)
	CheckSecret(secret string, hashedSecret []byte) error
}

func NewBcryptX(bcryptConfig config.BcryptConfig) BcryptX {
	return internal.NewBcryptX(bcryptConfig)
}

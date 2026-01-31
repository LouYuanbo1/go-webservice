package gormx

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/gormx/internal"
	"gorm.io/gorm"
)

type GormXTx interface {
	Exec(ctx context.Context, fn func(ctx context.Context) error) error
}

func NewGormXTx(db *gorm.DB) GormXTx {
	return internal.NewGormXTx(db)
}

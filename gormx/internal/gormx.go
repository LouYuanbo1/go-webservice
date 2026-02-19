package internal

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/gormx/model"
	"gorm.io/gorm"
)

type gormX[T any, ID comparable, PT model.PointerModel[T, ID]] struct {
	db *gorm.DB
}

func NewGormX[T any, ID comparable, PT model.PointerModel[T, ID]](db *gorm.DB) *gormX[T, ID, PT] {
	return &gormX[T, ID, PT]{db: db}
}

func (gx *gormX[T, ID, PT]) GetDBWithContext(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if !ok {
		return gx.db.WithContext(ctx)
	}
	return tx.WithContext(ctx)
}

func (gx *gormX[T, ID, PT]) InTransaction(ctx context.Context) bool {
	_, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	return ok
}

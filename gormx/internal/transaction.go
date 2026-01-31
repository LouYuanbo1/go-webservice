package internal

import (
	"context"

	"gorm.io/gorm"
)

type contextTxKey struct{}

type gormTx struct {
	db *gorm.DB
}

func NewGormXTx(db *gorm.DB) *gormTx {
	return &gormTx{db: db}
}

func (gt *gormTx) Exec(ctx context.Context, fn func(ctx context.Context) error) error {
	return gt.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, contextTxKey{}, tx)
		return fn(ctx)
	})
}

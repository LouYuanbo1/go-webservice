package gormx

import (
	"context"

	"gorm.io/gorm"
)

type GormXTx interface {
	/*
		你只需要将需要在事务中执行的数据库操作放入 fn 中即可,
		如果 fn 中返回了错误, 事务会回滚, 否则会提交事务.

		You just need to put the database operations you want to execute in the transaction in fn.
		If fn returns an error, the transaction will be rolled back, otherwise it will be committed.
	*/
	Exec(ctx context.Context, fn func(ctx context.Context) error) error
}

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

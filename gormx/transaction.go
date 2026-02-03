package gormx

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/gormx/internal"
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

func NewGormXTx(db *gorm.DB) GormXTx {
	return internal.NewGormXTx(db)
}

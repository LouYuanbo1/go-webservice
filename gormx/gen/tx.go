package gen

import (
	"context"

	"gorm.io/gorm"
)

type (
// contextTxKey struct{}
)

type Tx interface {
	//Exec(ctx context.Context, fn func(ctx context.Context) error) error
	Exec(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error
}

func NewTx(db *gorm.DB) Tx {
	return &tx{db: db}
}

type tx struct {
	db *gorm.DB
}

/*
	func (t *tx) Exec(ctx context.Context, fn func(ctx context.Context) error) error {
		return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			ctx = context.WithValue(ctx, contextTxKey{}, tx)
			return fn(ctx)
		})
	}
*/

func (t *tx) Exec(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, tx)
	})
}

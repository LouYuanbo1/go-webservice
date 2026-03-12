package gormx

import (
	"context"

	"gorm.io/gorm"
)

type (
// contextTxKey struct{}
)

type Conn interface {
	Session
	//Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	Transaction(ctx context.Context, fn func(ctx context.Context, s Session) error) error
}

func NewConn(db *gorm.DB) *conn {
	return &conn{db: db}
}

type conn struct {
	db *gorm.DB
}

/*
func (c *conn) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, contextTxKey{}, tx)
		return fn(ctx)
	})
}
*/

func (c *conn) Transaction(ctx context.Context, fn func(ctx context.Context, s Session) error) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		s := NewSession(tx)
		return fn(ctx, s)
	})
}

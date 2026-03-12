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
	//Exec(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error
	Exec(ctx context.Context, fn func(ctx context.Context, sf *SessionFactory) error) error
}

func NewTx(db *gorm.DB) Tx {
	return &tx{db: db}
}

type tx struct {
	db *gorm.DB
}

/*
选择1:直接暴露tx,破坏封装性
func (t *tx) Exec(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, tx)
	})
}
*/

/*
选择2:使用ctx隐式传递tx,Go官方不建议的方法,
func (t *tx) Exec(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, contextTxKey{}, tx)
		return fn(ctx)
	})
}
*/

/*
选择3:创建泛型创造工程,Go语言的泛型无法实现这一点

Go语言脆弱的泛型无法实现这一点

type SessionFactory interface {
	NewSession[T any, ID comparable, PT PointerModel[T, ID]]() Session[T, ID, PT]
}

// sessionFactory 实现了SessionFactory接口
type sessionFactory struct {
	db *gorm.DB // 仅内部持有，不对外暴露
}

// NewSession 创建绑定事务的Session实例（核心）
func (f *sessionFactory) NewSession[T any, ID comparable, PT PointerModel[T, ID]]() Session[T, ID, PT] {
	// 创建绑定事务的Session，底层使用事务tx
	return &sessionImpl[T, ID, PT]{
		db: f.tx, // 直接使用事务tx，无需从context获取
	}
}
*/

/*
选择4:使用包级泛型函数创建,详情查看session_factory.go:NewSessionFromFactory
*/

func (t *tx) Exec(ctx context.Context, fn func(ctx context.Context, sf *SessionFactory) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sf := &SessionFactory{db: tx}
		return fn(ctx, sf)
	})
}

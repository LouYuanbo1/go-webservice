package gormx

import (
	"context"

	"gorm.io/gorm"
)

/*
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


func (t *tx) Exec(ctx context.Context, fn func(ctx context.Context, sf *SessionFactory) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sf := &SessionFactory{db: tx}
		return fn(ctx, sf)
	})
}



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
选择4:使用包级泛型函数创建:
type SessionFactory struct {
	db *gorm.DB // 仅内部持有，不对外暴露
}

func NewSessionFromFactory[T any, ID comparable, PT PointerModel[T, ID]](sf *SessionFactory) Session[T, ID, PT] {
	return NewSession[T, ID, PT](sf.db)
}
*/

type TypedDB struct {
	db *gorm.DB
}

func NewTypedDB(db *gorm.DB) *TypedDB {
	return &TypedDB{db: db}
}

func (tdb *TypedDB) Transaction(ctx context.Context, fn func(ctx context.Context, tx *TypedDB) error) error {
	return tdb.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, &TypedDB{db: tx})
	})
}

func GetSession[T any, ID comparable, PT PointerModel[T, ID]](tdb *TypedDB) TypedSession[T, ID, PT] {
	return NewTypedSession[T, ID, PT](tdb.db)
}

/*
Transaction 开启事务，回调内会得到一个事务专用的 TypedDB 实例
// 非事务查询
type UserRepo struct {
	tdb *TypedDB
}

func NewUserRepo(tdb *TypedDB) *UserRepo {
	return &UserRepo{tdb: tdb}
}

func (ur *UserRepo) GetUser(ctx context.Context, id int64) (*User, error) {
    sess := GetSession[User, int64, *User](ur.tdb)
    var user User
    err := sess.GetByID(ctx, &user, id)
    return &user, err
}

// 跨表事务
func (ur *UserRepo) CreateOrderAndUpdateUser(ctx context.Context, order *Order, userID int64) error {
    return ur.tdb.Transaction(ctx, func(ctx context.Context, tx *TypedDB) error {
        // 在事务内创建 session，所有操作自动使用事务 DB
        orderSess := GetSession[Order, int64, *Order](tx)
        userSess := GetSession[User, int64, *User](tx)

        if err := orderSess.Create(ctx, order); err != nil {
            return err
        }
        return userSess.UpdateByID(ctx, userID, &User{Status: "active"})
    })
}
*/

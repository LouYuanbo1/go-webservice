package gormx

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/breaker"
	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
)

type DB struct {
	exec       *Executor
	brk        breaker.Breaker
	acceptable func(err error) bool // 自定义忽略错误：如记录不存在不算异常
}

// Option 用来自定义熔断器、关闭熔断、自定义acceptable
type Option func(db *DB)

// WithBreaker 自定义熔断器
func WithBreaker(brk breaker.Breaker) Option {
	return func(db *DB) {
		db.brk = brk
	}
}

// WithAcceptable 自定义错误白名单
func WithAcceptable(acc func(err error) bool) Option {
	return func(db *DB) {
		db.acceptable = acc
	}
}

func defaultAcceptable(err error) bool {
	return errorx.In(err, gorm.ErrRecordNotFound, gorm.ErrInvalidTransaction)
}

func NewDB(db *gorm.DB, opts ...Option) *DB {
	xdb := &DB{
		exec:       NewExecutor(db),
		brk:        breaker.NewBreaker(),
		acceptable: defaultAcceptable,
	}
	// 应用自定义配置
	for _, opt := range opts {
		opt(xdb)
	}
	return xdb
}

func (db *DB) Exec(ctx context.Context, fn func(gormDB *gorm.DB) error) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.Exec(ctx, fn)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrExecFailed,
				"gormx",
				"Exec db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) Create[T any, PT PointerModel[T]](ctx context.Context, model PT, opts ...ConflictOption) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.Create(ctx, model, opts...)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrCreateFailed,
				"gormx",
				"Create db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) CreateInBatches[T any, PT PointerModel[T]](ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.CreateInBatches(ctx, models, batchSize, opts...)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrCreateFailed,
				"gormx",
				"CreateInBatches db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) GetByID[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest PT, id ID) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.GetByID(ctx, dest, id)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				"GetByID db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) GetByFilter[T any, PT PointerModel[T]](ctx context.Context, dest PT, filter PT) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.GetByFilter(ctx, dest, filter)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				"GetByStructFilter db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) FindByIDs[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest *[]T, ids []ID, opts ...OrderOption) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.FindByIDs(ctx, dest, ids, opts...)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				"FindByIDs db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) FindByFilter[T any, PT PointerModel[T]](ctx context.Context, dest *[]T, filter PT, opts ...OrderOption) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.FindByFilter(ctx, dest, filter, opts...)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				"FindByStructFilter db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) FindByPage[T any, PT PointerModel[T]](ctx context.Context, dest *[]T, page, pageSize int, opts ...OrderOption) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.FindByPage(ctx, dest, page, pageSize, opts...)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				"FindByPage db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) FindByCursor[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest *[]T, cursor ID, limit int) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.FindByCursor(ctx, dest, cursor, limit)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				"FindByCursor db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) FindInBatches[T any, PT PointerModel[T]](
	ctx context.Context,
	batchSize int,
	callback func(ctx context.Context, tx *DB, batch int, models *[]T) error,
) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.FindInBatches(ctx, batchSize, callback)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				"FindInBatches db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) Update[T any, PT PointerModel[T]](ctx context.Context, updateData PT) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.Update(ctx, updateData)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrUpdateFailed,
				"gormx",
				"Update db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) UpdatesByFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT, updateData PT) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.UpdatesByFilter(ctx, filter, updateData)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrUpdateFailed,
				"gormx",
				"UpdatesByStructFilter db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) DeleteByID[T any, PT PointerModel[T], ID comparable](ctx context.Context, id ID) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.DeleteByID[T](ctx, id)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrDeleteFailed,
				"gormx",
				"DeleteByID db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) DeleteByIDs[T any, PT PointerModel[T], ID comparable](ctx context.Context, ids ...ID) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.DeleteByIDs[T](ctx, ids...)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrDeleteFailed,
				"gormx",
				"DeleteByIDs db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) DeleteByFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.DeleteByFilter(ctx, filter)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrDeleteFailed,
				"gormx",
				"DeleteByStructFilter db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) Transaction(ctx context.Context, fn func(ctx context.Context, tx *Tx) error) error {
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.exec.getDBWithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return fn(ctx, &Tx{exec: NewExecutor(tx)})
		})
	}, db.acceptable)
	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrTransactionFailed,
				"gormx",
				"Transaction db breaker open",
				err,
			)
		}
		return err
	}
	return nil
}

package gormx

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/breaker"
	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DB struct {
	exec       *Executor
	brk        breaker.Breaker
	acceptable func(err error) bool
}

type Option func(*DB)

func WithBreaker(brk breaker.Breaker) Option {
	return func(db *DB) { db.brk = brk }
}

func WithAcceptable(acc func(err error) bool) Option {
	return func(db *DB) { db.acceptable = acc }
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
	for _, opt := range opts {
		opt(xdb)
	}
	return xdb
}

func (db *DB) do(ctx context.Context, op string, fn func(*Executor) error) (err error) {
	ctx, span := startSpan(ctx, op)
	defer func() {
		endSpan(span, err)
	}()

	err = db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return fn(db.exec)
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				getBreakerError(op),
				"gormx",
				op,
				err,
			)
		}
		return err
	}
	return nil
}

func getBreakerError(op string) error {
	switch op {
	case "Create", "CreateInBatches":
		return ErrCreateFailed
	case "First":
		return ErrFirstFailed
	case "Find", "FindInBatches":
		return ErrFindFailed
	case "Count":
		return ErrCountFailed
	case "Update":
		return ErrUpdateFailed
	case "Updates":
		return ErrUpdatesFailed
	case "Delete":
		return ErrDeleteFailed
	default:
		return ErrNoRowsAffected
	}
}

func (db *DB) Build(fn func(tx *gorm.DB) *gorm.DB) *DB {
	return &DB{
		exec:       db.exec.Build(fn),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Model[T any, PT PointerModel[T]](model PT) *DB {
	return &DB{
		exec:       db.exec.Model(model),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Table(tableName string, args ...any) *DB {
	return &DB{
		exec:       db.exec.Table(tableName, args...),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Raw(query string, args ...any) *DB {
	return &DB{
		exec:       db.exec.Raw(query, args...),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Clauses(conds ...clause.Expression) *DB {
	return &DB{
		exec:       db.exec.Clauses(conds...),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Select(query any, args ...any) *DB {
	return &DB{
		exec:       db.exec.Select(query, args...),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Where(query any, args ...any) *DB {
	return &DB{
		exec:       db.exec.Where(query, args...),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) StructFilter[T any, PT PointerModel[T]](filter PT) *DB {
	return &DB{
		exec:       db.exec.StructFilter(filter),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) MapFilter(filter map[string]any) *DB {
	return &DB{
		exec:       db.exec.MapFilter(filter),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Order(order any) *DB {
	return &DB{
		exec:       db.exec.Order(order),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) OrderByColumn(clause clause.OrderByColumn) *DB {
	return &DB{
		exec:       db.exec.OrderByColumn(clause),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) OrderBy(clause clause.OrderBy) *DB {
	return &DB{
		exec:       db.exec.OrderBy(clause),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Joins(query string, args ...any) *DB {
	return &DB{
		exec:       db.exec.Joins(query, args...),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) InnerJoins(query string, args ...any) *DB {
	return &DB{
		exec:       db.exec.InnerJoins(query, args...),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Limit(limit int) *DB {
	return &DB{
		exec:       db.exec.Limit(limit),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Offset(offset int) *DB {
	return &DB{
		exec:       db.exec.Offset(offset),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Unscoped() *DB {
	return &DB{
		exec:       db.exec.Unscoped(),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Omit(columns ...string) *DB {
	return &DB{
		exec:       db.exec.Omit(columns...),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Group(name string) *DB {
	return &DB{
		exec:       db.exec.Group(name),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Having(query any, args ...any) *DB {
	return &DB{
		exec:       db.exec.Having(query, args...),
		brk:        db.brk,
		acceptable: db.acceptable,
	}
}

func (db *DB) Create[T any, PT PointerModel[T]](ctx context.Context, model PT) error {
	return db.do(ctx, "Create", func(exec *Executor) error {
		return exec.Create(ctx, model)
	})
}

func (db *DB) CreateInBatches[T any](ctx context.Context, models *[]T, batchSize int) error {
	return db.do(ctx, "CreateInBatches", func(exec *Executor) error {
		return exec.CreateInBatches(ctx, models, batchSize)
	})
}

func (db *DB) First[T any, PT PointerModel[T]](ctx context.Context, dest PT, conds ...any) error {
	return db.do(ctx, "First", func(exec *Executor) error {
		return exec.First(ctx, dest, conds...)
	})
}

func (db *DB) Scan[T any](ctx context.Context, dest *T) error {
	return db.do(ctx, "Scan", func(exec *Executor) error {
		return exec.Scan(ctx, dest)
	})
}

func (db *DB) Find[T any](ctx context.Context, dest *[]T, conds ...any) error {
	return db.do(ctx, "Find", func(exec *Executor) error {
		return exec.Find(ctx, dest, conds...)
	})
}

func (db *DB) FindInBatches[T any](ctx context.Context, batchSize int, fn func(tx *Executor, batch int, dest *[]T) error) error {
	return db.do(ctx, "FindInBatches", func(exec *Executor) error {
		return exec.FindInBatches(ctx, batchSize, fn)
	})
}

func (db *DB) Count(ctx context.Context, count *int64) error {
	return db.do(ctx, "Count", func(exec *Executor) error {
		return exec.Count(ctx, count)
	})
}

func (db *DB) Update(ctx context.Context, column string, value any) error {
	return db.do(ctx, "Update", func(exec *Executor) error {
		return exec.Update(ctx, column, value)
	})
}

func (db *DB) Updates[T any, PT PointerModel[T]](ctx context.Context, updateData PT) error {
	return db.do(ctx, "Updates", func(exec *Executor) error {
		return exec.Updates(ctx, updateData)
	})
}

func (db *DB) Delete[T any, PT PointerModel[T]](ctx context.Context, dest PT, conds ...any) error {
	return db.do(ctx, "Delete", func(exec *Executor) error {
		return exec.Delete(ctx, dest, conds...)
	})
}

func (db *DB) Transaction(ctx context.Context, fn func(tx *Executor) error) error {
	return db.do(ctx, "Transaction", func(exec *Executor) error {
		return exec.Transaction(ctx, fn)
	})
}

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

/*
	type Product struct {
		ID uint64 `gorm:"primaryKey"`
		ProductCode string `gorm:"unique"`
		ImageURL    string
	}

Clauses 用于添加自定义的 SQL 子句。
一般常用:
Clauses(

	clause.OnConflict{
		Columns: []clause.Column{
			{Name: "product_code"},
		},
		UpdateAll: true,
	},

)

Clauses(

	clause.OnConflict{
		Columns: []clause.Column{
			{Name: "product_code"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"image_url"}),
	},

)
*/
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

/*
Exec 执行原始 SQL 语句
*/
func (db *DB) Exec(ctx context.Context, sql string, values ...any) error {
	return db.do(ctx, "Exec", func(exec *Executor) error {
		return exec.Exec(ctx, sql, values...)
	})
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

/*
Delete（删除）：GORM 为了防止误删全表，
对结构体参数的处理极其严格——它只会检查结构体中的主键字段（即 ID）。
因为你的 ID 是 uint64，默认值为 0，GORM 认为你没有指定主键，属于“无条件的批量删除”，
于是触发安全机制，直接报错 WHERE conditions required，完全忽略了其他字段。

传入结构体没有主键会怎样？
会直接报错（就是你遇到的这种情况）。
即便结构体里有其他字段，GORM 也不会拿它们当 WHERE 条件，而是坚决拒绝执行，
除非你显式使用 Where 或 Unscoped（不推荐）绕过安全限制。

	type Product struct {
		ID uint64 `gorm:"primaryKey"`
		ProductCode string `gorm:"unique"`
		ImageURL    string
	}

✅ 方案一：使用 Where 链式调用（最推荐，语义最清晰）
go
// 注意：Delete 传入空结构体或 &Product{} 均可

	if err := db.Where("product_code = ?", code).Delete(ctx, &model.Product{}); err != nil {
	    return fmt.Errorf("删除产品失败: %w", err)
	}

✅ 方案二：直接使用 Delete 的第二个参数传条件（简洁写法）
go
// 直接在主方法后面跟 SQL 条件

	if err := db.Delete(ctx, &model.Product{}, "product_code = ?", code); err != nil {
	    return fmt.Errorf("删除产品失败: %w", err)
	}

❌ 错误写法（就是你现在的写法，必须改掉）
go
// 这样写 ProductCode 永远不会生效，因为主键 ID=0
db.Delete(ctx, &model.Product{ProductCode: code})
*/
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

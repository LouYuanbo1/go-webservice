package gormx

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Executor struct {
	db *gorm.DB
}

func NewExecutor(db *gorm.DB) *Executor {
	return &Executor{db: db}
}

func (e *Executor) Build(fn func(tx *gorm.DB) *gorm.DB) *Executor {
	nextDB := fn(e.db)
	return NewExecutor(nextDB)
}

func (e *Executor) Model[T any, PT PointerModel[T]](model PT) *Executor {
	return NewExecutor(e.db.Model(model))
}

/*
Table 表操作

典型使用场景：
1. 动态表名（表名含占位符）
go
db.Table("? as users", db.Raw("SELECT * FROM user_info WHERE age > ?", 18))
args 接收 db.Raw(...) 返回的表达式，最终拼成 (SELECT * FROM user_info WHERE age > 18) as users。
2. 带有参数的表名函数（如数据库函数）
go
db.Table("get_user_table(?)", userID)
适用于调用数据库函数返回表名，参数通过 ? 传递。
3. 直接传入子查询
go
subQuery := db.Table("orders").Select("user_id").Where("status = ?", "active")
db.Table("(?) as o", subQuery).Where("o.total > ?", 100)
args 接收子查询的 *gorm.DB 对象，GORM 会将其编译为 SQL 并作为子句。
*/

func (e *Executor) Table(tableName string, args ...any) *Executor {
	return NewExecutor(e.db.Table(tableName, args...))
}

func (e *Executor) Raw(query string, args ...any) *Executor {
	return NewExecutor(e.db.Raw(query, args...))
}

func (e *Executor) Clauses(conds ...clause.Expression) *Executor {
	return NewExecutor(e.db.Clauses(conds...))
}

// Select 指定要查询或更新的字段，后续操作（如 Find、Updates）将仅作用于这些字段,这种指定会解除结构体默认忽略0值的行为
func (e *Executor) Select(query any, args ...any) *Executor {
	return NewExecutor(e.db.Select(query, args...))
}

func (e *Executor) Where(query any, args ...any) *Executor {
	return NewExecutor(e.db.Where(query, args...))
}

func (e *Executor) StructFilter[T any, PT PointerModel[T]](filter PT) *Executor {
	return NewExecutor(e.db.Where(filter))
}

func (e *Executor) MapFilter(filter map[string]any) *Executor {
	return NewExecutor(e.db.Where(filter))
}

func (e *Executor) Order(order any) *Executor {
	return NewExecutor(e.db.Order(order))
}

func (e *Executor) OrderByColumn(clause clause.OrderByColumn) *Executor {
	return NewExecutor(e.db.Order(clause))
}

func (e *Executor) OrderBy(clause clause.OrderBy) *Executor {
	return NewExecutor(e.db.Order(clause))
}

func (e *Executor) Joins(query string, args ...any) *Executor {
	return NewExecutor(e.db.Joins(query, args...))
}

func (e *Executor) InnerJoins(query string, args ...any) *Executor {
	return NewExecutor(e.db.InnerJoins(query, args...))
}

func (e *Executor) Limit(limit int) *Executor {
	return NewExecutor(e.db.Limit(limit))
}

func (e *Executor) Offset(offset int) *Executor {
	return NewExecutor(e.db.Offset(offset))
}

func (e *Executor) Unscoped() *Executor {
	return NewExecutor(e.db.Unscoped())
}

func (e *Executor) Omit(columns ...string) *Executor {
	return NewExecutor(e.db.Omit(columns...))
}

func (e *Executor) Group(name string) *Executor {
	return NewExecutor(e.db.Group(name))
}

func (e *Executor) Having(query any, args ...any) *Executor {
	return NewExecutor(e.db.Having(query, args...))
}

func (e *Executor) Create[T any, PT PointerModel[T]](ctx context.Context, model PT) error {
	prefix := "Create"
	if model == nil {
		return errorx.New(ErrInvalidModel, "gormx", prefix, nil)
	}

	result := e.db.WithContext(ctx).Create(model)

	if result.Error != nil {
		return errorx.NewWithDetails(
			ErrCreateFailed,
			"gormx",
			prefix,
			result.Statement.Table,
			result.Error,
		)
	}
	return nil
}
func (e *Executor) CreateInBatches[T any](ctx context.Context, models *[]T, batchSize int) error {
	prefix := "CreateInBatches"
	if models == nil {
		return errorx.New(ErrInvalidModel, "gormx", prefix, nil)
	}

	result := e.db.WithContext(ctx).CreateInBatches(models, batchSize)

	if result.Error != nil {
		return errorx.NewWithDetails(
			ErrCreateFailed,
			"gormx",
			prefix,
			result.Statement.Table,
			result.Error,
		)
	}
	return nil
}

func (e *Executor) First[T any, PT PointerModel[T]](ctx context.Context, dest PT, conds ...any) error {
	prefix := "First"
	if dest == nil {
		return errorx.New(ErrInvalidModel, "gormx", prefix, nil)
	}

	result := e.db.WithContext(ctx).First(dest, conds...)

	if result.Error != nil {
		return errorx.NewWithDetails(ErrFirstFailed,
			"gormx",
			prefix,
			result.Statement.Table,
			result.Error,
		)

	}
	return nil
}

func (e *Executor) Scan[T any](ctx context.Context, dest *T) error {
	prefix := "Scan"
	if dest == nil {
		return errorx.New(ErrInvalidModel, "gormx", prefix, nil)
	}

	result := e.db.WithContext(ctx).Scan(dest)

	if result.Error != nil {
		return errorx.NewWithDetails(ErrScanFailed,
			"gormx",
			prefix,
			result.Statement.Table,
			result.Error,
		)

	}
	return nil
}

func (e *Executor) Find[T any](ctx context.Context, dest *[]T, conds ...any) error {
	prefix := "Find"
	if dest == nil {
		return errorx.New(ErrInvalidModel, "gormx", prefix, nil)
	}

	result := e.db.WithContext(ctx).Find(dest, conds...)

	if result.Error != nil {
		return errorx.NewWithDetails(ErrFindFailed,
			"gormx",
			prefix,
			result.Statement.Table,
			result.Error,
		)

	}
	return nil
}

func (e *Executor) FindInBatches[T any](ctx context.Context, batchSize int, fn func(tx *Executor, batch int, dest *[]T) error) error {
	prefix := "FindInBatches"

	dest := make([]T, batchSize)
	result := e.db.WithContext(ctx).FindInBatches(&dest, batchSize, func(tx *gorm.DB, batch int) error {
		return fn(NewExecutor(tx), batch, &dest)
	})

	if result.Error != nil {
		return errorx.NewWithDetails(
			ErrFindFailed,
			"gormx",
			prefix,
			result.Statement.Table,
			result.Error,
		)
	}
	return nil
}

func (e *Executor) Count(ctx context.Context, count *int64) error {
	prefix := "Count"
	if count == nil {
		return errorx.New(ErrInvalidModel, "gormx", prefix, nil)
	}

	result := e.db.WithContext(ctx).Count(count)

	if result.Error != nil {
		return errorx.NewWithDetails(
			ErrCountFailed,
			"gormx",
			prefix,
			result.Statement.Table,
			result.Error,
		)
	}
	return nil
}

// Update 更新单个字段（value 用 any 不可避免，但比 map 安全得多）
func (e *Executor) Update(ctx context.Context, column string, value any) error {
	prefix := "Update"

	result := e.db.WithContext(ctx).Update(column, value)

	if result.Error != nil {
		return errorx.NewWithDetails(
			ErrUpdateFailed,
			"gormx",
			prefix,
			result.Statement.Table,
			result.Error,
		)
	}
	return nil
}

func (e *Executor) Updates[T any, PT PointerModel[T]](ctx context.Context, updateData PT) error {
	prefix := "Updates"
	if updateData == nil {
		return errorx.New(ErrInvalidModel, "gormx", prefix, nil)
	}

	result := e.db.WithContext(ctx).Updates(updateData)

	if result.Error != nil {
		return errorx.NewWithDetails(ErrUpdatesFailed,
			"gormx",
			prefix,
			result.Statement.Table,
			result.Error,
		)
	}
	return nil
}

func (e *Executor) Delete[T any, PT PointerModel[T]](ctx context.Context, dest PT, conds ...any) error {
	prefix := "Delete"
	if dest == nil {
		return errorx.New(ErrInvalidModel, "gormx", prefix, nil)
	}

	result := e.db.WithContext(ctx).Delete(dest, conds...)

	if result.Error != nil {
		return errorx.NewWithDetails(ErrDeleteFailed,
			"gormx",
			prefix,
			result.Statement.Table,
			result.Error,
		)
	}
	return nil
}

func (e *Executor) Transaction(ctx context.Context, fn func(tx *Executor) error) error {
	return e.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewExecutor(tx))
	})
}

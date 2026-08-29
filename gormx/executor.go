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

/*
Exec 执行原始 SQL 语句,注意:使用Exec时,前方的查询条件会被忽略
*/
func (e *Executor) Exec(ctx context.Context, sql string, values ...any) error {
	prefix := "Exec"

	result := e.db.WithContext(ctx).Exec(sql, values...)

	if result.Error != nil {
		return errorx.NewWithDetails(
			ErrExecFailed,
			"gormx",
			prefix,
			result.Statement.Table,
			result.Error,
		)
	}
	return nil
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

/*
Pluck 从查询结果中提取指定字段的值,并将其存储到 dest 中
*/
func (e *Executor) Pluck[T any](ctx context.Context, column string, dest *T) error {
	prefix := "Pluck"
	if dest == nil {
		return errorx.New(ErrInvalidModel, "gormx", prefix, nil)
	}

	result := e.db.WithContext(ctx).Pluck(column, dest)

	if result.Error != nil {
		return errorx.NewWithDetails(ErrPluckFailed,
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

/*
UpdatesByStruct 使用结构体字段更新,忽略0值
*/
func (e *Executor) UpdatesByStruct[T any, PT PointerModel[T]](ctx context.Context, filter PT, updateData PT) error {
	prefix := "UpdatesByStruct"
	if updateData == nil {
		return errorx.New(ErrInvalidModel, "gormx", prefix, nil)
	}
	if filter == nil {
		return errorx.New(ErrInvalidFilter, "gormx", prefix, nil)
	}

	result := e.db.WithContext(ctx).Where(filter).Updates(updateData)

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

/*
UpdatesByMap 使用 map 字段更新,不忽略0值
*/
func (e *Executor) UpdatesByMap[T any, PT PointerModel[T]](ctx context.Context, filter map[string]any, updateData PT) error {
	prefix := "UpdatesByMap"
	if updateData == nil {
		return errorx.New(ErrInvalidModel, "gormx", prefix, nil)
	}
	if len(filter) == 0 {
		return errorx.New(ErrInvalidFilter, "gormx", prefix, nil)
	}

	result := e.db.WithContext(ctx).Where(filter).Updates(updateData)

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

/*
DeleteByStruct 使用结构体字段删除,忽略0值
*/
func (e *Executor) DeleteByStruct[T any, PT PointerModel[T]](ctx context.Context, filter PT) error {
	prefix := "DeleteByStruct"
	if filter == nil {
		return errorx.New(ErrInvalidFilter, "gormx", prefix, nil)
	}

	result := e.db.WithContext(ctx).Where(filter).Delete(new(T))

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

/*
DeleteByMap 使用 map 字段删除,不忽略0值,注意此函数需要显式模型类型
*/
func (e *Executor) DeleteByMap[T any](ctx context.Context, filter map[string]any) error {
	prefix := "DeleteByMap"
	if len(filter) == 0 {
		return errorx.New(ErrInvalidFilter, "gormx", prefix, nil)
	}

	// 核心：使用 new(T) 获取空模型，既获取表名，又保留 GORM 特性
	result := e.db.WithContext(ctx).
		Where(filter).
		Delete(new(T))

	if result.Error != nil {
		return errorx.NewWithDetails(ErrDeleteFailed,
			"gormx",
			prefix,
			result.Statement.Table, // 这里拿到的表名就是 GORM 解析后的真实表名
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

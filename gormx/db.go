package gormx

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
)

type DB struct {
	gdb *gorm.DB
}

func NewDB(db *gorm.DB) *DB {
	return &DB{gdb: db}
}

func (db *DB) GetDBWithContext(ctx context.Context) *gorm.DB {
	return db.gdb.WithContext(ctx)
}

func (db *DB) Create[T any, PT PointerModel[T]](ctx context.Context, model PT, opts ...ConflictOption) error {
	prefix := "Create"
	if model == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidModel)
		return nil
	}

	// 1. 构建基础 DB 对象
	gormDB := db.GetDBWithContext(ctx)

	// 2. 有冲突选项时，添加 ON CONFLICT 子句
	if len(opts) > 0 {
		clauseConflict, err := db.clauseOnConflictBuilder(opts...)
		if err != nil {
			return errorx.New(
				ErrInvalidOnConflictClause,
				"gormx",
				"Create",
				err,
			)
		}
		gormDB = gormDB.Clauses(clauseConflict)
		prefix = "Create(Upsert)"
	}

	// 3. 统一执行 Create
	result := gormDB.Create(model)
	if result.Error != nil {
		return errorx.New(
			ErrCreateFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}

	// 4. 统一处理 RowsAffected 日志
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

// 注意这里的models需要在调用时传入一个slice类型的参数
func (db *DB) CreateInBatches[T any, PT PointerModel[T]](ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) error {
	prefix := "CreateInBatches"
	// 参数校验
	if batchSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidBatchSize)
		return nil
	}
	if models == nil {
		// 空切片属于合法操作（0 行插入），静默成功更符合批量操作语义
		log.Printf("%s skipped: %s", prefix, WarnEmptyModelsSlice)
		return nil
	}

	// 1. 构建基础 DB 对象
	gormDB := db.GetDBWithContext(ctx)

	// 2. 有冲突选项时，添加 ON CONFLICT 子句
	if len(opts) > 0 {
		clauseConflict, err := db.clauseOnConflictBuilder(opts...)
		if err != nil {
			return errorx.New(
				ErrInvalidOnConflictClause,
				"gormx",
				"Create",
				err,
			)
		}
		gormDB = gormDB.Clauses(clauseConflict)
		prefix = "CreateInBatches(Upsert)"
	}

	result := gormDB.CreateInBatches(models, batchSize)
	if result.Error != nil {
		return errorx.New(
			ErrCreateFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (db *DB) GetByID[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest PT, id ID) error {
	prefix := "GetByID"
	if IsZero(id) {
		log.Printf("%s failed : %s", prefix, WarnInvalidID)
		return nil
	}

	result := db.GetDBWithContext(ctx).
		First(dest, id)
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (db *DB) GetByStructFilter[T any, PT PointerModel[T]](ctx context.Context, dest PT, filter PT) error {
	prefix := "GetByStructFilter"
	if filter == nil {
		log.Printf("%s failed: %s", prefix, WarnInvalidFilter)
		return nil
	}

	result := db.GetDBWithContext(ctx).
		Where(filter).
		First(dest)
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (db *DB) FindByIDs[T any, ID comparable](ctx context.Context, dest *[]T, ids []ID, opts ...OrderOption) error {
	prefix := "FindByIDs"
	if len(ids) == 0 {
		log.Printf("%s failed : %s", prefix, WarnEmptyIDSlice)
		return nil
	}

	// 1. 构建基础 DB 对象
	gormDB := db.GetDBWithContext(ctx)

	// 2. 有排序选项时，添加 ORDER BY 子句
	if len(opts) > 0 {
		clauseOrder := db.clauseOrderBuilder(opts...)
		gormDB = gormDB.Order(clauseOrder)
		prefix = "FindByIDs(Order)"
	}

	result := gormDB.Find(dest, ids)
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (db *DB) FindByStructFilter[T any, PT PointerModel[T]](ctx context.Context, dest *[]T, filter PT, opts ...OrderOption) error {
	prefix := "FindByStructFilter"
	if filter == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidFilter)
		return nil
	}

	// 1. 构建基础 DB 对象
	gormDB := db.GetDBWithContext(ctx)

	// 2. 有排序选项时，添加 ORDER BY 子句
	if len(opts) > 0 {
		clauseOrder := db.clauseOrderBuilder(opts...)
		gormDB = gormDB.Order(clauseOrder)
		prefix = "FindByStructFilter(Order)"
	}

	result := gormDB.Where(filter).Find(dest)
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func getPrimaryKeyColumns[T any](db *gorm.DB) (string, error) {
	// 独立会话，DryRun 防止意外执行 SQL
	model := new(T)
	stmt := db.Session(&gorm.Session{DryRun: true}).Model(model).Statement
	// 显式解析 dest，填充 Schema
	if err := stmt.Parse(stmt.Model); err != nil {
		return "", fmt.Errorf("解析模型失败: %w", err)
	}
	if stmt.Schema == nil {
		return "", fmt.Errorf("无法解析模型 Schema")
	}
	fields := stmt.Schema.PrimaryFieldDBNames
	if len(fields) == 0 {
		return "", fmt.Errorf("模型未定义主键")
	}
	return strings.Join(fields, ", "), nil
}

func (db *DB) FindByPage[T any](ctx context.Context, dest *[]T, page, pageSize int, opts ...OrderOption) error {
	prefix := "FindByPage"
	if page <= 0 || pageSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidPageParams)
		return nil
	}

	// 1. 构建基础 DB 对象
	gormDB := db.GetDBWithContext(ctx).Model(dest)

	// 2. 有排序选项时，添加 ORDER BY 子句
	if len(opts) > 0 {
		clauseOrder := db.clauseOrderBuilder(opts...)
		gormDB = gormDB.Order(clauseOrder)
		prefix = "FindByPage(Order)"
	} else {
		// 自动获取主键排序
		pkColumns, err := getPrimaryKeyColumns[T](db.gdb)
		if err != nil {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s failed: get primary key", prefix),
				err,
			)
		}
		clauseOrder := fmt.Sprintf("%s ASC", pkColumns)
		gormDB = gormDB.Order(clauseOrder)
	}

	result := gormDB.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(dest)

	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (db *DB) FindByCursor[T any, ID comparable](ctx context.Context, dest *[]T, cursor ID, limit int) error {
	prefix := "FindByCursor"
	if limit <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidLimit)
		return nil
	}

	gormDB := db.GetDBWithContext(ctx).Model(dest)
	pkColumns, err := getPrimaryKeyColumns[T](db.gdb)
	if err != nil {
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			"FindByCursor: get primary key",
			err,
		)
	}
	result := gormDB.
		Where(fmt.Sprintf("%s > ?", pkColumns), cursor).
		Order(fmt.Sprintf("%s ASC", pkColumns)).
		Limit(limit).
		Find(dest)
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (db *DB) FindInBatches[T any](
	ctx context.Context,
	batchSize int,
	callback func(ctx context.Context, tx *DB, batch int, models *[]T) error,
) error {
	prefix := "FindInBatches"
	if batchSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidBatchSize)
		return nil
	}

	// 1. 构建基础 DB 对象
	gormDB := db.GetDBWithContext(ctx)

	dest := make([]T, 0, batchSize)

	result := gormDB.FindInBatches(&dest, batchSize, func(tx *gorm.DB, batch int) error {
		return callback(ctx, NewDB(tx), batch, &dest)
	})
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (db *DB) Update[T any, PT PointerModel[T]](ctx context.Context, updateData PT) error {
	prefix := "Update"
	if updateData == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidUpdateData)
		return nil
	}

	result := db.GetDBWithContext(ctx).
		Updates(updateData)
	if result.Error != nil {
		return errorx.New(
			ErrUpdateFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (db *DB) UpdatesByStructFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT, updateData PT) error {
	prefix := "UpdatesByStructFilter"
	if updateData == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidUpdateData)
		return nil
	}
	if filter == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidFilter)
		return nil
	}

	result := db.GetDBWithContext(ctx).
		Where(filter).
		Updates(updateData)
	if result.Error != nil {
		return errorx.New(
			ErrUpdateFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (db *DB) DeleteByID[T any, PT PointerModel[T], ID comparable](ctx context.Context, model PT, id ID) error {
	prefix := "DeleteByID"
	result := db.GetDBWithContext(ctx).
		Delete(model, id)
	if result.Error != nil {
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (db *DB) DeleteByIDs[T any, PT PointerModel[T], ID comparable](ctx context.Context, model PT, ids ...ID) error {
	prefix := "DeleteByIDs"
	if ids == nil {
		log.Printf("%s failed : %s", prefix, WarnEmptyIDSlice)
		return nil
	}

	result := db.GetDBWithContext(ctx).
		Delete(model, ids)
	if result.Error != nil {
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (db *DB) DeleteByStructFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT) error {
	prefix := "DeleteByStructFilter"
	if filter == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidFilter)
		return nil
	}

	result := db.GetDBWithContext(ctx).
		Where(filter).
		Delete(PT(new(T)))
	if result.Error != nil {
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (db *DB) Transaction(ctx context.Context, fn func(ctx context.Context, tx *DB) error) error {
	return db.GetDBWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, NewDB(tx))
	})
}

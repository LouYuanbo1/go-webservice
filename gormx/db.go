package gormx

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/LouYuanbo1/go-webservice/breaker"
	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
)

type DB struct {
	gdb        *gorm.DB
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
		gdb:        db,
		brk:        breaker.NewBreaker(),
		acceptable: defaultAcceptable,
	}
	// 应用自定义配置
	for _, opt := range opts {
		opt(xdb)
	}
	return xdb
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

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		gormDB := db.GetDBWithContext(ctx)
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

		result := gormDB.Create(model)
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrCreateFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}

		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}

		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrCreateFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) CreateInBatches[T any, PT PointerModel[T]](ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) error {
	prefix := "CreateInBatches"
	if batchSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidBatchSize)
		return nil
	}
	if models == nil {
		log.Printf("%s skipped: %s", prefix, WarnEmptyModelsSlice)
		return nil
	}

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		gormDB := db.GetDBWithContext(ctx)

		if len(opts) > 0 {
			clauseConflict, err := db.clauseOnConflictBuilder(opts...)
			if err != nil {
				return errorx.New(
					ErrInvalidOnConflictClause,
					"gormx",
					"CreateInBatches",
					err,
				)
			}
			gormDB = gormDB.Clauses(clauseConflict)
			prefix = "CreateInBatches(Upsert)"
		}

		result := gormDB.CreateInBatches(models, batchSize)
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrCreateFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrCreateFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) GetByID[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest PT, id ID) error {
	prefix := "GetByID"
	if IsZero(id) {
		log.Printf("%s failed : %s", prefix, WarnInvalidID)
		return nil
	}

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		result := db.GetDBWithContext(ctx).First(dest, id)
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) GetByStructFilter[T any, PT PointerModel[T]](ctx context.Context, dest PT, filter PT) error {
	prefix := "GetByStructFilter"
	if filter == nil {
		log.Printf("%s failed: %s", prefix, WarnInvalidFilter)
		return nil
	}

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		result := db.GetDBWithContext(ctx).Where(filter).First(dest)
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) FindByIDs[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest *[]T, ids []ID, opts ...OrderOption) error {
	prefix := "FindByIDs"
	if len(ids) == 0 {
		log.Printf("%s failed : %s", prefix, WarnEmptyIDSlice)
		return nil
	}

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		gormDB := db.GetDBWithContext(ctx).Model(PT(new(T)))

		if len(opts) > 0 {
			clauseOrder := db.clauseOrderBuilder(opts...)
			gormDB = gormDB.Order(clauseOrder)
			prefix = "FindByIDs(Order)"
		}

		result := gormDB.Find(dest, ids)
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) FindByStructFilter[T any, PT PointerModel[T]](ctx context.Context, dest *[]T, filter PT, opts ...OrderOption) error {
	prefix := "FindByStructFilter"
	if filter == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidFilter)
		return nil
	}

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		gormDB := db.GetDBWithContext(ctx)

		if len(opts) > 0 {
			clauseOrder := db.clauseOrderBuilder(opts...)
			gormDB = gormDB.Order(clauseOrder)
			prefix = "FindByStructFilter(Order)"
		}

		result := gormDB.Where(filter).Find(dest)
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
	}
	return nil
}

func getPrimaryKeyColumns[T any, PT PointerModel[T]](db *gorm.DB) (string, error) {
	stmt := db.Session(&gorm.Session{DryRun: true}).Model(PT(new(T))).Statement
	if err := stmt.Parse(stmt.Model); err != nil {
		return "", errorx.New(
			ErrParseModelFailed,
			"gormx",
			"getPrimaryKey",
			err,
		)
	}
	if stmt.Schema == nil {
		return "", errorx.New(
			ErrParseModelFailed,
			"gormx",
			"getPrimaryKey",
			nil,
		)
	}
	fields := stmt.Schema.PrimaryFieldDBNames
	if len(fields) == 0 {
		return "", errorx.New(
			ErrModelNoPrimaryKey,
			"gormx",
			"getPrimaryKey",
			nil,
		)
	}
	return strings.Join(fields, ", "), nil
}

func (db *DB) FindByPage[T any, PT PointerModel[T]](ctx context.Context, dest *[]T, page, pageSize int, opts ...OrderOption) error {
	prefix := "FindByPage"
	if page <= 0 || pageSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidPageParams)
		return nil
	}

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		gormDB := db.GetDBWithContext(ctx)

		if len(opts) > 0 {
			clauseOrder := db.clauseOrderBuilder(opts...)
			gormDB = gormDB.Order(clauseOrder)
			prefix = "FindByPage(Order)"
		} else {
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

		result := gormDB.Offset((page - 1) * pageSize).Limit(pageSize).Find(dest)
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) FindByCursor[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest *[]T, cursor ID, limit int) error {
	prefix := "FindByCursor"
	if limit <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidLimit)
		return nil
	}

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		gormDB := db.GetDBWithContext(ctx)
		pkColumns, err := getPrimaryKeyColumns[T](db.gdb)
		if err != nil {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				"FindByCursor: get primary key",
				err,
			)
		}

		result := gormDB.Where(fmt.Sprintf("%s > ?", pkColumns), cursor).Order(fmt.Sprintf("%s ASC", pkColumns)).Limit(limit).Find(dest)
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
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
	prefix := "FindInBatches"
	if batchSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidBatchSize)
		return nil
	}

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		gormDB := db.GetDBWithContext(ctx).Model(PT(new(T)))
		dest := make([]T, 0, batchSize)

		result := gormDB.FindInBatches(&dest, batchSize, func(tx *gorm.DB, batch int) error {
			return callback(ctx, NewDB(tx), batch, &dest)
		})
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) Update[T any, PT PointerModel[T]](ctx context.Context, updateData PT) error {
	prefix := "Update"
	if updateData == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidUpdateData)
		return nil
	}

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		result := db.GetDBWithContext(ctx).Updates(updateData)
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrUpdateFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrUpdateFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
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

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		result := db.GetDBWithContext(ctx).Where(filter).Updates(updateData)
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrUpdateFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrUpdateFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) DeleteByID[T any, PT PointerModel[T], ID comparable](ctx context.Context, id ID) error {
	prefix := "DeleteByID"

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		result := db.GetDBWithContext(ctx).Delete(PT(new(T)), id)
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrDeleteFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrDeleteFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) DeleteByIDs[T any, PT PointerModel[T], ID comparable](ctx context.Context, ids ...ID) error {
	prefix := "DeleteByIDs"
	if ids == nil {
		log.Printf("%s failed : %s", prefix, WarnEmptyIDSlice)
		return nil
	}

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		result := db.GetDBWithContext(ctx).Delete(PT(new(T)), ids)
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrDeleteFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrDeleteFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) DeleteByStructFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT) error {
	prefix := "DeleteByStructFilter"
	if filter == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidFilter)
		return nil
	}

	var tableName string
	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		result := db.GetDBWithContext(ctx).Where(filter).Delete(PT(new(T)))
		tableName = result.Statement.Table

		if result.Error != nil {
			return errorx.New(
				ErrDeleteFailed,
				"gormx",
				fmt.Sprintf("%s[%s]", prefix, tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("%s failed. table: %s, %s", prefix, tableName, WarnNoRowsAffected)
		}
		return nil
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrDeleteFailed,
				"gormx",
				fmt.Sprintf("%s[%s] db breaker open", prefix, tableName),
				err,
			)
		}
		return err
	}
	return nil
}

func (db *DB) Transaction(ctx context.Context, fn func(ctx context.Context, tx *DB) error) error {
	prefix := "Transaction"

	err := db.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
		return db.GetDBWithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return fn(ctx, NewDB(tx))
		})
	}, db.acceptable)

	if err != nil {
		if errorx.Is(err, breaker.ErrServiceUnavailable) {
			return errorx.New(
				ErrTransactionFailed,
				"gormx",
				fmt.Sprintf("%s db breaker open", prefix),
				err,
			)
		}
		return errorx.New(
			ErrTransactionFailed,
			"gormx",
			prefix,
			err,
		)
	}
	return nil
}

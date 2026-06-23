package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
)

type Tx struct {
	gdb *gorm.DB
}

func (tx *Tx) Exec(ctx context.Context, fn func(gormDB *gorm.DB) error) error {
	prefix := "Exec"
	gormDB := tx.gdb.WithContext(ctx) // 在这里构造，不暴露给外层
	if err := fn(gormDB); err != nil {
		return errorx.New(
			ErrExecFailed,
			"gormx",
			prefix,
			err,
		)
	}
	return nil
}

func (tx *Tx) Create[T any, PT PointerModel[T]](ctx context.Context, model PT, opts ...ConflictOption) error {
	prefix := "Create"
	if model == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidModel)
		return nil
	}

	gormDB := tx.gdb.WithContext(ctx)
	if len(opts) > 0 {
		clauseConflict, err := clauseOnConflictBuilder(opts...)
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
	tableName := result.Statement.Table

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
}

func (tx *Tx) CreateInBatches[T any, PT PointerModel[T]](ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) error {
	prefix := "CreateInBatches"
	if batchSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidBatchSize)
		return nil
	}
	if models == nil {
		log.Printf("%s skipped: %s", prefix, WarnEmptyModelsSlice)
		return nil
	}

	gormDB := tx.gdb.WithContext(ctx)
	if len(opts) > 0 {
		clauseConflict, err := clauseOnConflictBuilder(opts...)
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
	tableName := result.Statement.Table

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
}

func (tx *Tx) GetByID[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest PT, id ID) error {
	prefix := "GetByID"
	if IsZero(id) {
		log.Printf("%s failed : %s", prefix, WarnInvalidID)
		return nil
	}

	result := tx.gdb.WithContext(ctx).First(dest, id)
	tableName := result.Statement.Table

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
}

func (tx *Tx) GetByStructFilter[T any, PT PointerModel[T]](ctx context.Context, dest PT, filter PT) error {
	prefix := "GetByStructFilter"
	if filter == nil {
		log.Printf("%s failed: %s", prefix, WarnInvalidFilter)
		return nil
	}

	result := tx.gdb.WithContext(ctx).Where(filter).First(dest)
	tableName := result.Statement.Table

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

}

func (tx *Tx) FindByIDs[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest *[]T, ids []ID, opts ...OrderOption) error {
	prefix := "FindByIDs"
	if len(ids) == 0 {
		log.Printf("%s failed : %s", prefix, WarnEmptyIDSlice)
		return nil
	}

	gormDB := tx.gdb.WithContext(ctx).Model(PT(new(T)))
	if len(opts) > 0 {
		clauseOrder := clauseOrderBuilder(opts...)
		gormDB = gormDB.Order(clauseOrder)
		prefix = "FindByIDs(Order)"
	}

	result := gormDB.Find(dest, ids)
	tableName := result.Statement.Table

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
}

func (tx *Tx) FindByStructFilter[T any, PT PointerModel[T]](ctx context.Context, dest *[]T, filter PT, opts ...OrderOption) error {
	prefix := "FindByStructFilter"
	if filter == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidFilter)
		return nil
	}

	gormDB := tx.gdb.WithContext(ctx)
	if len(opts) > 0 {
		clauseOrder := clauseOrderBuilder(opts...)
		gormDB = gormDB.Order(clauseOrder)
		prefix = "FindByStructFilter(Order)"
	}

	result := gormDB.Where(filter).Find(dest)
	tableName := result.Statement.Table

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
}

func (tx *Tx) FindByPage[T any, PT PointerModel[T]](ctx context.Context, dest *[]T, page, pageSize int, opts ...OrderOption) error {
	prefix := "FindByPage"
	if page <= 0 || pageSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidPageParams)
		return nil
	}

	gormDB := tx.gdb.WithContext(ctx)
	if len(opts) > 0 {
		clauseOrder := clauseOrderBuilder(opts...)
		gormDB = gormDB.Order(clauseOrder)
		prefix = "FindByPage(Order)"
	} else {
		pkColumns, err := getPrimaryKeyColumns[T](tx.gdb)
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
	tableName := result.Statement.Table

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
}

func (tx *Tx) FindByCursor[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest *[]T, cursor ID, limit int) error {
	prefix := "FindByCursor"
	if limit <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidLimit)
		return nil
	}

	gormDB := tx.gdb.WithContext(ctx)
	pkColumns, err := getPrimaryKeyColumns[T](tx.gdb)
	if err != nil {
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			"FindByCursor: get primary key",
			err,
		)
	}

	result := gormDB.Where(fmt.Sprintf("%s > ?", pkColumns), cursor).Order(fmt.Sprintf("%s ASC", pkColumns)).Limit(limit).Find(dest)
	tableName := result.Statement.Table

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
}

func (tx *Tx) FindInBatches[T any, PT PointerModel[T]](
	ctx context.Context,
	batchSize int,
	callback func(ctx context.Context, tx *DB, batch int, models *[]T) error,
) error {
	prefix := "FindInBatches"
	if batchSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidBatchSize)
		return nil
	}

	gormDB := tx.gdb.WithContext(ctx).Model(PT(new(T)))
	dest := make([]T, 0, batchSize)

	result := gormDB.FindInBatches(&dest, batchSize, func(tx *gorm.DB, batch int) error {
		return callback(ctx, NewDB(tx), batch, &dest)
	})
	tableName := result.Statement.Table

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
}

func (tx *Tx) Update[T any, PT PointerModel[T]](ctx context.Context, updateData PT) error {
	prefix := "Update"
	if updateData == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidUpdateData)
		return nil
	}

	result := tx.gdb.WithContext(ctx).Updates(updateData)
	tableName := result.Statement.Table

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
}

func (tx *Tx) UpdatesByStructFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT, updateData PT) error {
	prefix := "UpdatesByStructFilter"
	if updateData == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidUpdateData)
		return nil
	}
	if filter == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidFilter)
		return nil
	}

	result := tx.gdb.WithContext(ctx).Where(filter).Updates(updateData)
	tableName := result.Statement.Table

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
}

func (tx *Tx) DeleteByID[T any, PT PointerModel[T], ID comparable](ctx context.Context, id ID) error {
	prefix := "DeleteByID"

	result := tx.gdb.WithContext(ctx).Delete(PT(new(T)), id)
	tableName := result.Statement.Table

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
}

func (tx *Tx) DeleteByIDs[T any, PT PointerModel[T], ID comparable](ctx context.Context, ids ...ID) error {
	prefix := "DeleteByIDs"
	if ids == nil {
		log.Printf("%s failed : %s", prefix, WarnEmptyIDSlice)
		return nil
	}

	result := tx.gdb.WithContext(ctx).Delete(PT(new(T)), ids)
	tableName := result.Statement.Table

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
}

func (tx *Tx) DeleteByStructFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT) error {
	prefix := "DeleteByStructFilter"
	if filter == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidFilter)
		return nil
	}

	result := tx.gdb.WithContext(ctx).Where(filter).Delete(PT(new(T)))
	tableName := result.Statement.Table

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
}

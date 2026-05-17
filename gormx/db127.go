package gormx

/*
import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
)

type DB127 struct {
	db *gorm.DB
}

func NewDB127(db *gorm.DB) *DB127 {
	return &DB127{db: db}
}

func (xdb *DB127) GetDBWithContext(ctx context.Context) *gorm.DB {
	return xdb.db.WithContext(ctx)
}

func (xdb *DB127) Create[T any, PT PointerModel[T]](ctx context.Context, model PT, opts ...ConflictOption) error {
	if model == nil {
		log.Printf("create failed : %s", WarnInvalidModel)
		return nil
	}

	var result *gorm.DB
	// 应用冲突选项
	if len(opts) == 0 {
		result = xdb.GetDBWithContext(ctx).
			Create(model)
		if result.Error != nil {
			return errorx.New(
				ErrCreateFailed,
				"gormx",
				fmt.Sprintf("Create[%s]", result.Statement.Table),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("create failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
		}
		return nil
	}

	clauseConflict, err := s.clauseOnConflictBuilder(opts...)
	if err != nil {
		return errorx.New(
			ErrInvalidOnConflictClause,
			"gormx",
			"Create",
			err,
		)
	}

	result = xdb.GetDBWithContext(ctx).
		Clauses(clauseConflict).
		Create(model)
	if result.Error != nil {
		return errorx.New(
			ErrCreateFailed,
			"gormx",
			fmt.Sprintf("Create(Upsert)[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("create(upsert) failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

// 注意这里的models需要在调用时传入一个slice类型的参数
func (xdb *DB127) CreateInBatches[T any, PT PointerModel[T]](ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) error {
	// 参数校验
	if batchSize <= 0 {
		log.Printf("create in batches failed : %s", WarnInvalidBatchSize)
		return nil
	}
	if len(models) == 0 {
		// 空切片属于合法操作（0 行插入），静默成功更符合批量操作语义
		log.Printf("skipped create in batches: %s", WarnEmptyModelsSlice)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = xdb.GetDBWithContext(ctx).
			CreateInBatches(models, batchSize)
		if result.Error != nil {
			log.Printf("create in batches failed. table: %s, error: %v", result.Statement.Table, result.Error)
			return errorx.New(
				ErrCreateFailed,
				"gormx",
				fmt.Sprintf("CreateInBatches[%s]", result.Statement.Table),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("create in batches failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
		}
		return nil
	}

	// 应用冲突选项
	clauseConflict, err := s.clauseOnConflictBuilder(opts...)
	if err != nil {
		return errorx.New(
			ErrInvalidOnConflictClause,
			"gormx",
			"CreateInBatches",
			err,
		)
	}

	result = xdb.GetDBWithContext(ctx).
		Clauses(clauseConflict).
		CreateInBatches(models, batchSize)
	if result.Error != nil {
		log.Printf("create(upsert) in batches failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrCreateFailed,
			"gormx",
			fmt.Sprintf("CreateInBatches(Upsert)[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("create in batches failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) GetByID[T any, PT PointerModel[T]](ctx context.Context, dest PT, id comparable) error {
	if IsZero(id) {
		log.Printf("get by id failed : %s", WarnInvalidID)
		return nil
	}

	result := xdb.GetDBWithContext(ctx).
		First(dest, id)
	if result.Error != nil {
		log.Printf("get by id failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("GetByID[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by id failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) GetByStructFilter[T any, PT PointerModel[T]](ctx context.Context, dest PT, filter PT) error {
	if filter == nil {
		log.Printf("get by struct filter failed : %s", WarnInvalidFilter)
		return nil
	}

	result := xdb.GetDBWithContext(ctx).
		Where(filter).
		First(dest)
	if result.Error != nil {
		log.Printf("get by struct filter failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("GetByStructFilter[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by struct filter failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) FindByIDs[T any, PT PointerModel[T]](ctx context.Context, dest *[]PT, ids []comparable, opts ...OrderOption) error {
	if len(ids) == 0 {
		log.Printf("find by ids failed : %s", WarnEmptyIDSlice)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = xdb.GetDBWithContext(ctx).
			Find(dest, ids)
		if result.Error != nil {
			log.Printf("find by ids failed. table: %s, error: %v", result.Statement.Table, result.Error)
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindByIDs[%s]", result.Statement.Table),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by ids failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
		}
		return nil
	}

	clauseOrder := s.clauseOrderBuilder(opts...)

	result = xdb.GetDBWithContext(ctx).
		Order(clauseOrder).
		Find(dest, ids)
	if result.Error != nil {
		log.Printf("find by ids failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindByIDs(Order)[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by ids (order) failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) FindByStructFilter[T any, PT PointerModel[T]](ctx context.Context, dest *[]PT, filter PT, opts ...OrderOption) error {
	if filter == nil {
		log.Printf("find by struct filter failed : %s", WarnInvalidFilter)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = xdb.GetDBWithContext(ctx).
			Where(filter).
			Find(dest)
		if result.Error != nil {
			log.Printf("find by struct filter failed. table: %s, error: %v", result.Statement.Table, result.Error)
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindByStructFilter[%s]", result.Statement.Table),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by struct filter failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
		}
		return nil
	}

	clauseOrder := s.clauseOrderBuilder(opts...)

	result = xdb.GetDBWithContext(ctx).
		Where(filter).
		Order(clauseOrder).
		Find(dest)
	if result.Error != nil {
		log.Printf("find by struct filter (order) failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindByStructFilter(Order)[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by struct filter (order) failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) FindByPage[T any, PT PointerModel[T]](ctx context.Context, dest *[]PT, page, pageSize int, opts ...OrderOption) error {
	if page <= 0 || pageSize <= 0 {
		log.Printf("find by page %d, pageSize %d failed : %s", page, pageSize, WarnInvalidPageParams)
		return nil
	}

	var result *gorm.DB
	model := PT(new(T))

	if len(opts) == 0 {
		result = xdb.GetDBWithContext(ctx).
			Order(fmt.Sprintf("%s ASC", model.PrimaryKey())).
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(dest)
		if result.Error != nil {
			log.Printf("find by page %d, pageSize %d failed. table: %s, error: %v", page, pageSize, result.Statement.Table, result.Error)
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindByPage[%s]", result.Statement.Table),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by page %d, pageSize %d failed. table: %s, %s", page, pageSize, result.Statement.Table, WarnNoRowsAffected)
		}
		return nil
	}

	clauseOrder := s.clauseOrderBuilder(opts...)

	result = xdb.GetDBWithContext(ctx).
		Order(clauseOrder).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(dest)
	if result.Error != nil {
		log.Printf("find by page %d, pageSize %d (order) failed. table: %s, error: %v", page, pageSize, result.Statement.Table, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindByPage(Order)[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by page %d, pageSize %d (order) failed. table: %s, %s", page, pageSize, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) FindByCursor[T any, PT PointerModel[T]](ctx context.Context, dest *[]PT, cursor comparable, limit int) error {
	if limit <= 0 {
		log.Printf("find by cursor failed : %s", WarnInvalidLimit)
		return nil
	}

	model := PT(new(T))

	result := xdb.GetDBWithContext(ctx).
		Where(fmt.Sprintf("%s > ?", model.PrimaryKey()), cursor).
		Order(fmt.Sprintf("%s ASC", model.PrimaryKey())).
		Limit(limit).
		Find(dest)
	if result.Error != nil {
		log.Printf("find by cursor %v, limit %d failed. table: %s, error: %v", cursor, limit, result.Statement.Table, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindByCursor[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by cursor %v, limit %d failed. table: %s, %s", cursor, limit, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) FindInBatches[T any, PT PointerModel[T]](
	ctx context.Context,
	batchSize int,
	callback func(ctx context.Context, tx *DB127, batch int, models PT) error,
	opts ...OrderOption,
) error {
	if batchSize <= 0 {
		log.Printf("find in batches failed : %s", WarnInvalidBatchSize)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = xdb.GetDBWithContext(ctx).
			FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
				return callback(ctx, NewDB127(tx), batch, dest)
			})
		if result.Error != nil {
			log.Printf("find in batches failed. table: %s, error: %v", result.Statement.Table, result.Error)
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindInBatches[%s]", result.Statement.Table),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find in batches failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
		}
		return nil
	}

	clauseOrder := s.clauseOrderBuilder(opts...)

	result = s.GetDBWithContext(ctx).
		Order(clauseOrder).
		FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
			return callback(ctx, NewDB127(tx), batch, dest)
		})
	if result.Error != nil {
		log.Printf("find in batches (order) failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindInBatches(Order)[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find in batches (order) failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) FindInBatchesByStructFilter[T any, PT PointerModel[T]](
	ctx context.Context,
	filter PT,
	batchSize int,
	callback func(ctx context.Context, tx *DB127, batch int, models PT) error,
	opts ...OrderOption,
) error {
	if batchSize <= 0 {
		log.Printf("find in batches by struct filter failed : %s", WarnInvalidBatchSize)
		return nil
	}
	if filter == nil {
		log.Printf("find in batches by struct filter failed : %s", WarnInvalidFilter)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = xdb.GetDBWithContext(ctx).
			Where(filter).
			FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
				return callback(ctx, NewDB127(tx), batch, dest)
			})
		if result.Error != nil {
			log.Printf("find in batches by struct filter failed. table: %s, error: %v", result.Statement.Table, result.Error)
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindInBatchesByStructFilter[%s]", result.Statement.Table),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find in batches by struct filter failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
		}
		return nil
	}

	clauseOrder := s.clauseOrderBuilder(opts...)

	result = xdb.GetDBWithContext(ctx).
		Where(filter).
		Order(clauseOrder).
		FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
			//ctx = context.WithValue(ctx, contextTxKey{}, tx)
			return callback(ctx, NewDB127(tx), batch, dest)
		})
	if result.Error != nil {
		log.Printf("find in batches by struct filter (order) failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindInBatchesByStructFilter(Order)[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find in batches by struct filter (order) failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) Update[T any, PT PointerModel[T]](ctx context.Context, updateData PT) error {
	if updateData == nil {
		log.Printf("update failed : %s", WarnInvalidUpdateData)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Updates(updateData)
	if result.Error != nil {
		log.Printf("update failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrUpdateFailed,
			"gormx",
			fmt.Sprintf("Update[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("update failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) UpdatesByStructFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT, updateData PT) error {
	if updateData == nil {
		log.Printf("updates by struct filter failed : %s", WarnInvalidUpdateData)
		return nil
	}
	if filter == nil {
		log.Printf("updates by struct filter failed : %s", WarnInvalidFilter)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Where(filter).
		Updates(updateData)
	if result.Error != nil {
		log.Printf("updates by struct filter %v failed. table: %s error: %v", filter, result.Statement.Table, result.Error)
		return errorx.New(
			ErrUpdateFailed,
			"gormx",
			fmt.Sprintf("UpdatesByStructFilter[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("updates by struct filter failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) DeleteByID[T any, PT PointerModel[T]](ctx context.Context, model PT, id comparable) error {
	result := xdb.GetDBWithContext(ctx).
		Delete(model, id)
	if result.Error != nil {
		log.Printf("delete by id %v failed. table: %s, error: %v", id, result.Statement.Table, result.Error)
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("DeleteByID[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by id failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) DeleteByIDs[T any, PT PointerModel[T]](ctx context.Context, model PT, ids ...comparable) error {
	if ids == nil {
		log.Printf("delete by ids failed : %s", WarnEmptyIDSlice)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Delete(model, ids)
	if result.Error != nil {
		log.Printf("delete by ids %v failed. table: %s error: %v", ids, result.Statement.Table, result.Error)
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("DeleteByIDs[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by ids failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) DeleteByStructFilter[T any, PT PointerModel[T]](ctx context.Context, model PT, filter PT) error {
	if filter == nil {
		log.Printf("delete by struct filter failed : %s", WarnInvalidFilter)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Where(filter).
		Delete(model)
	if result.Error != nil {
		log.Printf("delete by struct filter %v failed. table: %s error: %v", filter, result.Statement.Table, result.Error)
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("DeleteByStructFilter[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by struct filter failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (xdb *DB127) Transaction(ctx context.Context, fn func(ctx context.Context, tx *DB127) error) error {
	return xdb.GetDBWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, NewDB127(tx))
	})
}
*/

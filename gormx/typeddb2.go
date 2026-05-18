package gormx

/*
import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
)

type TypedDB struct {
	DB
}

func NewTypedDB(db DB) *TypedDB {
	return &TypedDB{DB: db}
}

func (tdb *TypedDB) Create[T any, PT PointerModel[T]](ctx context.Context, model PT, opts ...ConflictOption) error {
	return tdb.DB.Create(ctx, model, opts...)
}

// 注意这里的models需要在调用时传入一个slice类型的参数
func (tdb *TypedDB) CreateInBatches[T any, PT PointerModel[T]](ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) error {
	if len(models) == 0 {
		// 空切片属于合法操作（0 行插入），静默成功更符合批量操作语义
		log.Printf("skipped create in batches: %s", WarnEmptyModelsSlice)
		return nil
	}
	return tdb.DB.CreateInBatches(ctx, models, batchSize, opts...)
}

func (tdb *TypedDB) GetByID[T any, PT PointerModel[T]](ctx context.Context, dest PT, id comparable) error {
	if IsZero(id) {
		log.Printf("get by id failed : %s", WarnInvalidID)
		return nil
	}
	return tdb.DB.GetByID(ctx, dest, id)
}

func (tdb *TypedDB) GetByStructFilter[T any, PT PointerModel[T]](ctx context.Context, dest PT, filter PT) error {
	return tdb.DB.GetByStructFilter(ctx, dest, filter)
}

func (tdb *TypedDB) FindByIDs[T any, PT PointerModel[T]](ctx context.Context, dest *[]PT, ids []comparable, opts ...OrderOption) error {
	if len(ids) == 0 {
		log.Printf("find by ids failed : %s", WarnEmptyIDSlice)
		return nil
	}
	for _, id := range ids {
		if IsZero(id) {
			log.Printf("find by ids failed, index: %v : %s", id, WarnInvalidID)
			return nil
		}
	}
	return tdb.DB.FindByIDs(ctx, dest, ids, opts...)
}

func (tdb *TypedDB) FindByStructFilter[T any, PT PointerModel[T]](ctx context.Context, dest *[]PT, filter PT, opts ...OrderOption) error {
	return tdb.DB.FindByStructFilter(ctx, dest, filter, opts...)
}

func (tdb *TypedDB) FindByPage[T any, PT PointerModel[T]](ctx context.Context, dest *[]PT, page, pageSize int, opts ...OrderOption) error {
	model := PT(new(T))
	return tdb.DB.FindByPage(ctx, dest, model.PrimaryKey(), page, pageSize, opts...)
}

func (tdb *TypedDB) FindByCursor[T any, PT PointerModel[T]](ctx context.Context, dest *[]PT, cursor comparable, limit int) (newCursor comparable, hasMore bool, err error) {
	if limit <= 0 {
		log.Printf("find by cursor failed : %s", WarnInvalidLimit)
		return cursor, false, nil
	}
	model := PT(new(T))

	err = tdb.DB.FindByCursor(ctx, dest, model.PrimaryKey(), cursor, limit+1)
	if err != nil {
		return cursor, false, err
	}
	hasMore = len(*dest) > limit
	if hasMore {
		*dest = (*dest)[:limit]
	}
	newCursor = cursor
	if len(*dest) > 0 {
		newCursor = (*dest)[len(*dest)-1].GetID()
	}
	return newCursor, hasMore, nil
}

func (tdb *TypedDB) FindInBatches[T any, PT PointerModel[T]](
	ctx context.Context,
	batchSize int,
	callback func(ctx context.Context, tx *TypedDB, batch int, models PT) error,
	opts ...OrderOption,
) error {
	if batchSize <= 0 {
		log.Printf("find in batches failed : %s", WarnInvalidBatchSize)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = tdb.GetDBWithContext(ctx).
			FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
				return callback(ctx, NewTypedDB(tx), batch, dest)
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
			return callback(ctx, NewTypedDB(tx), batch, dest)
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

func (tdb *TypedDB) FindInBatchesByStructFilter[T any, PT PointerModel[T]](
	ctx context.Context,
	filter PT,
	batchSize int,
	callback func(ctx context.Context, tx *TypedDB, batch int, models PT) error,
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
		result = tdb.GetDBWithContext(ctx).
			Where(filter).
			FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
				return callback(ctx, NewTypedDB(tx), batch, dest)
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

	result = tdb.GetDBWithContext(ctx).
		Where(filter).
		Order(clauseOrder).
		FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
			//ctx = context.WithValue(ctx, contextTxKey{}, tx)
			return callback(ctx, NewTypedDB(tx), batch, dest)
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

func (tdb *TypedDB) Update[T any, PT PointerModel[T]](ctx context.Context, updateData PT) error {
	return tdb.DB.Update(ctx, updateData)
}

func (tdb *TypedDB) UpdatesByStructFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT, updateData PT) error {
	return tdb.DB.UpdatesByStructFilter(ctx, filter, updateData)
}


func (tdb *TypedDB) DeleteByID[T any, PT PointerModel[T]](ctx context.Context, model PT, id comparable) error {
	if IsZero(id) {
		log.Printf("delete by id failed : %s", WarnInvalidID)
		return nil
	}

	model := PT(new(T))
	return tdb.DB.DeleteByID(ctx, model, id)
}

func (tdb *TypedDB) DeleteByIDs[T any, PT PointerModel[T]](ctx context.Context, model PT, ids ...comparable) error {
	if len(ids) == 0 {
		log.Printf("delete by ids failed : %s", WarnEmptyIDSlice)
		return nil
	}
	for _, id := range ids {
		if IsZero(id) {
			log.Printf("delete by ids failed, index: %v : %s", id, WarnInvalidID)
			return nil
		}
	}

	model := PT(new(T))
	return tdb.DB.DeleteByIDs(ctx, model, ids...)
}

func (tdb *TypedDB) DeleteByStructFilter[T any, PT PointerModel[T]](ctx context.Context, model PT, filter PT) error {
	model := PT(new(T))
	return tdb.DB.DeleteByStructFilter(ctx, model, filter)
}

func (tdb *TypedDB) Transaction(ctx context.Context, fn func(ctx context.Context, tx *TypedDB) error) error {
	return tdb.GetDBWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, NewTypedDB(tx))
	})
}
*/

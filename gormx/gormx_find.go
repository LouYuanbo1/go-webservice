package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
)

func (gx *gormX[T, ID, PT]) FindByIDs(ctx context.Context, ids []ID, opts ...OrderOption) ([]PT, error) {
	if len(ids) == 0 {
		log.Printf("find by ids failed : %s", WarnEmptyIDSlice)
		return nil, nil
	}
	for _, id := range ids {
		if IsZero(id) {
			log.Printf("find by ids failed, index: %v : %s", id, WarnInvalidID)
			return nil, nil
		}
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()
	ptrs := make([]PT, 0, len(ids))
	var result *gorm.DB

	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			Find(&ptrs, ids)
		if result.Error != nil {
			log.Printf("find by ids failed. table: %s, error: %v", tableName, result.Error)
			return nil, errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindByIDs[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by ids failed. table: %s, %s", tableName, WarnNoRowsAffected)
		}
		return ptrs, nil
	}

	clauseOrder := gx.clauseOrderBuilder(opts...)

	result = gx.GetDBWithContext(ctx).
		Order(clauseOrder).
		Find(&ptrs, ids)
	if result.Error != nil {
		log.Printf("find by ids failed. table: %s, error: %v", tableName, result.Error)
		return nil, errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindByIDs(Order)[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by ids (order) failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return ptrs, nil
}

func (gx *gormX[T, ID, PT]) FindByStructFilter(ctx context.Context, filter PT, opts ...OrderOption) ([]PT, error) {
	if filter == nil {
		log.Printf("find by struct filter failed : %s", WarnInvalidFilter)
		return nil, nil
	}
	// 预分配内存，避免后续扩容导致的性能问题, 初始容量设为50,后续可能可以设置
	ptrs := make([]PT, 0, 50)
	tableName := filter.TableName()
	var result *gorm.DB

	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			Where(filter).
			Find(&ptrs)
		if result.Error != nil {
			log.Printf("find by struct filter failed. table: %s, error: %v", tableName, result.Error)
			return nil, errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindByStructFilter[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by struct filter failed. table: %s, %s", tableName, WarnNoRowsAffected)
		}
		return ptrs, nil
	}

	clauseOrder := gx.clauseOrderBuilder(opts...)

	result = gx.GetDBWithContext(ctx).
		Where(filter).
		Order(clauseOrder).
		Find(&ptrs)
	if result.Error != nil {
		log.Printf("find by struct filter (order) failed. table: %s, error: %v", tableName, result.Error)
		return nil, errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindByStructFilter(Order)[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by ids (order) failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return ptrs, nil
}

func (gx *gormX[T, ID, PT]) FindByMapFilter(ctx context.Context, filter map[string]any, opts ...OrderOption) ([]PT, error) {
	if len(filter) == 0 {
		log.Printf("find by map filter failed : %s", WarnInvalidFilter)
		return nil, nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()
	ptrs := make([]PT, 0, 50)
	var result *gorm.DB

	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			Where(filter).
			Find(&ptrs)
		if result.Error != nil {
			log.Printf("find by map filter failed. table: %s, error: %v", tableName, result.Error)
			return nil, errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindByMapFilter[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by map filter failed. table: %s, %s", tableName, WarnNoRowsAffected)
		}
		return ptrs, nil
	}

	clauseOrder := gx.clauseOrderBuilder(opts...)

	result = gx.GetDBWithContext(ctx).
		Where(filter).
		Order(clauseOrder).
		Find(&ptrs)
	if result.Error != nil {
		log.Printf("find by map filter failed. table: %s, error: %v", tableName, result.Error)
		return nil, errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindByMapFilter(Order)[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by map filter (order) failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return ptrs, nil
}

func (gx *gormX[T, ID, PT]) FindByPage(ctx context.Context, page, pageSize int, opts ...OrderOption) ([]PT, error) {
	if page <= 0 || pageSize <= 0 {
		log.Printf("find by page %d, pageSize %d failed : %s", page, pageSize, WarnInvalidPageParams)
		return nil, nil
	}

	var model T
	ptr := PT(&model)
	primaryKey := ptr.PrimaryKey()
	tableName := ptr.TableName()
	ptrs := make([]PT, 0, pageSize)
	var result *gorm.DB

	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			Order(fmt.Sprintf("%s ASC", primaryKey)).
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&ptrs)
		if result.Error != nil {
			log.Printf("find by page %d, pageSize %d failed. table: %s, error: %v", page, pageSize, tableName, result.Error)
			return nil, errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindByPage[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by page %d, pageSize %d failed. table: %s, %s", page, pageSize, tableName, WarnNoRowsAffected)
		}
		return ptrs, nil
	}

	clauseOrder := gx.clauseOrderBuilder(opts...)

	result = gx.GetDBWithContext(ctx).
		Order(clauseOrder).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&ptrs)
	if result.Error != nil {
		log.Printf("find by page %d, pageSize %d (order) failed. table: %s, error: %v", page, pageSize, tableName, result.Error)
		return nil, errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindByPage(Order)[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by page %d, pageSize %d (order) failed. table: %s, %s", page, pageSize, tableName, WarnNoRowsAffected)
	}
	return ptrs, nil
}

func (gx *gormX[T, ID, PT]) FindByCursor(ctx context.Context, cursor ID, limit int) ([]PT, ID, bool, error) {
	if limit <= 0 {
		log.Printf("find by cursor failed : %s", WarnInvalidLimit)
		return nil, cursor, false, nil
	}
	if IsZero(cursor) {
		log.Printf("find by cursor failed : %s", WarnInvalidID)
		return nil, cursor, false, nil
	}

	var model T
	ptr := PT(&model)
	primaryKey := ptr.PrimaryKey()
	tableName := ptr.TableName()
	ptrs := make([]PT, 0, limit+1)

	result := gx.GetDBWithContext(ctx).
		Where(fmt.Sprintf("%s > ?", primaryKey), cursor).
		Order(fmt.Sprintf("%s ASC", primaryKey)).
		Limit(limit + 1).
		Find(&ptrs)
	if result.Error != nil {
		log.Printf("find by cursor %v, limit %d failed. table: %s, error: %v", cursor, limit, tableName, result.Error)
		return nil, cursor, false, errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindByCursor[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by cursor %v, limit %d failed. table: %s, %s", cursor, limit, tableName, WarnNoRowsAffected)
	}
	hasMore := len(ptrs) > limit
	if hasMore {
		ptrs = ptrs[:limit]
	}
	newCursor := cursor
	if len(ptrs) > 0 {
		newCursor = ptrs[len(ptrs)-1].GetID()
	}
	return ptrs, newCursor, hasMore, nil
}

func (gx *gormX[T, ID, PT]) FindInBatches(ctx context.Context, batchSize int, callback func(ctx context.Context, batch int, ptrModels []PT) error, opts ...OrderOption) error {
	if batchSize <= 0 {
		log.Printf("find in batches failed : %s", WarnInvalidBatchSize)
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()
	ptrs := make([]PT, 0, batchSize)
	var result *gorm.DB

	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			FindInBatches(&ptrs, batchSize, func(tx *gorm.DB, batch int) error {
				ctx = context.WithValue(ctx, contextTxKey{}, tx)
				return callback(ctx, batch, ptrs)
			})
		if result.Error != nil {
			log.Printf("find in batches failed. table: %s, error: %v", tableName, result.Error)
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindInBatches[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find in batches failed. table: %s, %s", tableName, WarnNoRowsAffected)
		}
		return nil
	}

	clauseOrder := gx.clauseOrderBuilder(opts...)

	result = gx.GetDBWithContext(ctx).
		Order(clauseOrder).
		FindInBatches(&ptrs, batchSize, func(tx *gorm.DB, batch int) error {
			ctx = context.WithValue(ctx, contextTxKey{}, tx)
			return callback(ctx, batch, ptrs)
		})
	if result.Error != nil {
		log.Printf("find in batches (order) failed. table: %s, error: %v", tableName, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindInBatches(Order)[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find in batches (order) failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) FindInBatchesByStructFilter(ctx context.Context, filter PT, batchSize int, callback func(ctx context.Context, batch int, ptrModels []PT) error, opts ...OrderOption) error {
	if filter == nil {
		log.Printf("find in batches by struct filter failed : %s", WarnInvalidFilter)
		return nil
	}
	if batchSize <= 0 {
		log.Printf("find in batches by struct filter failed : %s", WarnInvalidBatchSize)
		return nil
	}

	tableName := filter.TableName()
	ptrs := make([]PT, 0, batchSize)
	var result *gorm.DB

	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			Where(filter).
			FindInBatches(&ptrs, batchSize, func(tx *gorm.DB, batch int) error {
				ctx = context.WithValue(ctx, contextTxKey{}, tx)
				return callback(ctx, batch, ptrs)
			})
		if result.Error != nil {
			log.Printf("find in batches by struct filter failed. table: %s, error: %v", tableName, result.Error)
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindInBatchesByStructFilter[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find in batches by struct filter failed. table: %s, %s", tableName, WarnNoRowsAffected)
		}
		return nil
	}

	clauseOrder := gx.clauseOrderBuilder(opts...)

	result = gx.GetDBWithContext(ctx).
		Where(filter).
		Order(clauseOrder).
		FindInBatches(&ptrs, batchSize, func(tx *gorm.DB, batch int) error {
			ctx = context.WithValue(ctx, contextTxKey{}, tx)
			return callback(ctx, batch, ptrs)
		})
	if result.Error != nil {
		log.Printf("find in batches by struct filter (order) failed. table: %s, error: %v", tableName, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindInBatchesByStructFilter(Order)[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find in batches by struct filter (order) failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) FindInBatchesByMapFilter(ctx context.Context, filter map[string]any, batchSize int, callback func(ctx context.Context, batch int, ptrModels []PT) error, opts ...OrderOption) error {
	if len(filter) == 0 {
		log.Printf("find in batches by map filter failed : %s", WarnInvalidFilter)
		return nil
	}
	if batchSize <= 0 {
		log.Printf("find in batches by map filter failed : %s", WarnInvalidBatchSize)
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()
	ptrs := make([]PT, 0, batchSize)
	var result *gorm.DB

	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			Where(filter).
			FindInBatches(&ptrs, batchSize, func(tx *gorm.DB, batch int) error {
				ctx = context.WithValue(ctx, contextTxKey{}, tx)
				return callback(ctx, batch, ptrs)
			})
		if result.Error != nil {
			log.Printf("find in batches by map filter failed. table: %s, error: %v", tableName, result.Error)
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindInBatchesByMapFilter[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find in batches by map filter failed. table: %s, %s", tableName, WarnNoRowsAffected)
		}
		return nil
	}

	clauseOrder := gx.clauseOrderBuilder(opts...)

	result = gx.GetDBWithContext(ctx).
		Where(filter).
		Order(clauseOrder).
		FindInBatches(&ptrs, batchSize, func(tx *gorm.DB, batch int) error {
			ctx = context.WithValue(ctx, contextTxKey{}, tx)
			return callback(ctx, batch, ptrs)
		})
	if result.Error != nil {
		log.Printf("find in batches by map filter (order) failed. table: %s, error: %v", tableName, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindInBatchesByMapFilter(Order)[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find in batches by map filter (order) failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return nil
}

package gen

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/LouYuanbo1/go-webservice/gormx"
)

func (g *genSession[T, ID, PT]) FindByIDs(ctx context.Context, ids []ID, opts ...gormx.OrderOption) ([]PT, error) {
	if len(ids) == 0 {
		log.Printf("find by ids failed : %s", gormx.WarnEmptyIDSlice)
		return nil, nil
	}
	for _, id := range ids {
		if IsZero(id) {
			log.Printf("find by ids failed, index: %v : %s", id, gormx.WarnInvalidID)
			return nil, nil
		}
	}

	ptrs := make([]PT, 0, len(ids))
	/*
		tableName := ptr.TableName()
		ptrs := make([]PT, 0, len(ids))
		var result *gorm.DB

		if len(opts) == 0 {
			result = gx.GetDBWithContext(ctx).
				Find(&ptrs, ids)
			if result.Error != nil {
				log.Printf("find by ids failed. table: %s, error: %v", tableName, result.Error)
				return nil, errorx.New(
					gormx.ErrQueryFailed,
					"gormx",
					fmt.Sprintf("FindByIDs[%s]", tableName),
					result.Error,
				)
			}
			if result.RowsAffected == 0 {
				log.Printf("find by ids failed. table: %s, %s", tableName, gormx.WarnNoRowsAffected)
			}
			return ptrs, nil
		}

		clauseOrder := gormx.ClauseOrderBuilder(opts...)

		result = gx.GetDBWithContext(ctx).
			Order(clauseOrder).
			Find(&ptrs, ids)
		if result.Error != nil {
			log.Printf("find by ids failed. table: %s, error: %v", tableName, result.Error)
			return nil, errorx.New(
				gormx.ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindByIDs(Order)[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by ids (order) failed. table: %s, %s", tableName, gormx.WarnNoRowsAffected)
		}
	*/
	if err := g.Session.FindByIDs(ctx, &ptrs, ids, opts...); err != nil {
		return nil, err
	}
	return ptrs, nil
}

func (g *genSession[T, ID, PT]) FindByStructFilter(ctx context.Context, filter PT, opts ...gormx.OrderOption) ([]PT, error) {
	/*
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
	*/
	ptrs := make([]PT, 0, 50)
	if err := g.Session.FindByStructFilter(ctx, &ptrs, filter, opts...); err != nil {
		return nil, err
	}
	return ptrs, nil
}

func (g *genSession[T, ID, PT]) FindByMapFilter(ctx context.Context, filter map[string]any, opts ...gormx.OrderOption) ([]PT, error) {
	/*
		if len(filter) == 0 {
			log.Printf("find by map filter failed : %s", gormx.WarnInvalidFilter)
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
	*/
	ptrs := make([]PT, 0, 50)
	if err := g.Session.FindByMapFilter(ctx, &ptrs, filter, opts...); err != nil {
		return nil, err
	}
	return ptrs, nil
}

func (g *genSession[T, ID, PT]) FindByPage(ctx context.Context, page, pageSize int, opts ...gormx.OrderOption) ([]PT, error) {
	/*
		if page <= 0 || pageSize <= 0 {
			log.Printf("find by page %d, pageSize %d failed : %s", page, pageSize, gormx.WarnInvalidPageParams)
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
	*/
	var model T
	ptr := PT(&model)
	primaryKey := ptr.PrimaryKey()
	ptrs := make([]PT, 0, pageSize)
	if err := g.Session.FindByPage(ctx, &ptrs, primaryKey, page, pageSize, opts...); err != nil {
		return nil, err
	}
	return ptrs, nil
}

func (g *genSession[T, ID, PT]) FindByCursor(ctx context.Context, cursor ID, limit int) ([]PT, ID, bool, error) {
	if limit <= 0 {
		log.Printf("find by cursor failed : %s", gormx.WarnInvalidLimit)
		return nil, cursor, false, nil
	}
	if IsZero(cursor) {
		log.Printf("find by cursor failed : %s", gormx.WarnInvalidID)
		return nil, cursor, false, nil
	}

	var model T
	ptr := PT(&model)
	primaryKey := ptr.PrimaryKey()
	ptrs := make([]PT, 0, limit+1)
	/*
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
	*/
	err := g.Session.FindByCursor(ctx, &ptrs, primaryKey, cursor, limit)
	if err != nil {
		return nil, cursor, false, err
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

// 此方法与传统方法不兼容
func (g *genSession[T, ID, PT]) FindInBatches(
	ctx context.Context,
	batchSize int,
	callback func(ctx context.Context, batch int, models []PT) error,
	opts ...gormx.OrderOption,
) error {
	/*
		if batchSize <= 0 {
			log.Printf("find in batches failed : %s", gormx.WarnInvalidBatchSize)
			return nil
		}

		ptrs := make([]PT, 0, batchSize)
		var result *gorm.DB

		if len(opts) == 0 {
			result = g.GetDBWithContext(ctx).
				FindInBatches(&ptrs, batchSize, func(tx *gorm.DB, batch int) error {
					ctx = context.WithValue(ctx, contextTxKey{}, tx)
					return callback(ctx, batch, ptrs)
				})
			if result.Error != nil {
				log.Printf("find in batches failed. table: %s, error: %v", result.Statement.Table, result.Error)
				return errorx.New(
					gormx.ErrQueryFailed,
					"gormx",
					fmt.Sprintf("FindInBatches[%s]", result.Statement.Table),
					result.Error,
				)
			}
			if result.RowsAffected == 0 {
				log.Printf("find in batches failed. table: %s, %s", result.Statement.Table, gormx.WarnNoRowsAffected)
			}
			return nil
		}

		clauseOrder := g.ClauseOrderBuilder(opts...)

		result = g.GetDBWithContext(ctx).
			Order(clauseOrder).
			FindInBatches(&ptrs, batchSize, func(tx *gorm.DB, batch int) error {
				ctx = context.WithValue(ctx, contextTxKey{}, tx)
				return callback(ctx, batch, ptrs)
			})
		if result.Error != nil {
			log.Printf("find in batches (order) failed. table: %s, error: %v", result.Statement.Table, result.Error)
			return errorx.New(
				gormx.ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindInBatches(Order)[%s]", result.Statement.Table),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find in batches (order) failed. table: %s, %s", result.Statement.Table, gormx.WarnNoRowsAffected)
		}
	*/
	ptrs := make([]PT, 0, batchSize)
	err := g.Session.FindInBatches(ctx, &ptrs, batchSize,
		func(ctx context.Context, batch int, models any) error {
			typedModels, ok := models.([]PT)
			if !ok {
				return errorx.NewWithDetails(
					gormx.ErrInvalidTypeAssertion,
					"gormx",
					"FindInBatchesByStructFilter",
					fmt.Sprintf("unexpected type: %T", models),
					nil,
				)
			}
			return callback(ctx, batch, typedModels)
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (g *genSession[T, ID, PT]) FindInBatchesByStructFilter(
	ctx context.Context,
	filter PT,
	batchSize int,
	callback func(ctx context.Context, batch int, ptrModels []PT) error,
	opts ...gormx.OrderOption,
) error {
	/*
		if filter == nil {
			log.Printf("find in batches by struct filter failed : %s", gormx.WarnInvalidFilter)
			return nil
		}
		if batchSize <= 0 {
			log.Printf("find in batches by struct filter failed : %s", gormx.WarnInvalidBatchSize)
			return nil
		}

		ptrs := make([]PT, 0, batchSize)
		var result *gorm.DB

		if len(opts) == 0 {
			result = g.GetDBWithContext(ctx).
				Where(filter).
				FindInBatches(&ptrs, batchSize, func(tx *gorm.DB, batch int) error {
					ctx = context.WithValue(ctx, contextTxKey{}, tx)
					return callback(ctx, batch, ptrs)
				})
			if result.Error != nil {
				log.Printf("find in batches by struct filter failed. table: %s, error: %v", result.Statement.Table, result.Error)
				return errorx.New(
					gormx.ErrQueryFailed,
					"gormx",
					fmt.Sprintf("FindInBatchesByStructFilter[%s]", result.Statement.Table),
					result.Error,
				)
			}
			if result.RowsAffected == 0 {
				log.Printf("find in batches by struct filter failed. table: %s, %s",
					result.Statement.Table,
					gormx.WarnNoRowsAffected,
				)
			}
			return nil
		}

		clauseOrder := g.clauseOrderBuilder(opts...)

		result = g.GetDBWithContext(ctx).
			Where(filter).
			Order(clauseOrder).
			FindInBatches(&ptrs, batchSize, func(tx *gorm.DB, batch int) error {
				ctx = context.WithValue(ctx, contextTxKey{}, tx)
				return callback(ctx, batch, ptrs)
			})
		if result.Error != nil {
			log.Printf("find in batches by struct filter (order) failed. table: %s, error: %v", result.Statement.Table, result.Error)
			return errorx.New(
				gormx.ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindInBatchesByStructFilter(Order)[%s]", result.Statement.Table),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find in batches by struct filter (order) failed. table: %s, %s",
				result.Statement.Table,
				gormx.WarnNoRowsAffected,
			)
		}
	*/
	ptrs := make([]PT, 0, batchSize)
	err := g.Session.FindInBatchesByStructFilter(ctx, &ptrs, filter, batchSize,
		func(ctx context.Context, batch int, models any) error {
			typedModels, ok := models.([]PT)
			if !ok {
				return errorx.NewWithDetails(
					gormx.ErrInvalidTypeAssertion,
					"gormx",
					"FindInBatchesByStructFilter",
					fmt.Sprintf("unexpected type: %T", models),
					nil,
				)
			}
			return callback(ctx, batch, typedModels)
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (g *genSession[T, ID, PT]) FindInBatchesByMapFilter(
	ctx context.Context,
	filter map[string]any,
	batchSize int,
	callback func(ctx context.Context, batch int, ptrModels []PT) error,
	opts ...gormx.OrderOption,
) error {
	/*
		if len(filter) == 0 {
			log.Printf("find in batches by map filter failed : %s", gormx.WarnInvalidFilter)
			return nil
		}
		if batchSize <= 0 {
			log.Printf("find in batches by map filter failed : %s", gormx.WarnInvalidBatchSize)
			return nil
		}

		ptrs := make([]PT, 0, batchSize)
		var result *gorm.DB

		if len(opts) == 0 {
			result = g.GetDBWithContext(ctx).
				Where(filter).
				FindInBatches(&ptrs, batchSize, func(tx *gorm.DB, batch int) error {
					ctx = context.WithValue(ctx, contextTxKey{}, tx)
					return callback(ctx, batch, ptrs)
				})
			if result.Error != nil {
				log.Printf("find in batches by map filter failed. table: %s, error: %v", result.Statement.Table, result.Error)
				return errorx.New(
					gormx.ErrQueryFailed,
					"gormx",
					fmt.Sprintf("FindInBatchesByMapFilter[%s]", result.Statement.Table),
					result.Error,
				)
			}
			if result.RowsAffected == 0 {
				log.Printf("find in batches by map filter failed. table: %s, %s", result.Statement.Table, gormx.WarnNoRowsAffected)
			}
			return nil
		}

		clauseOrder := g.clauseOrderBuilder(opts...)

		result = g.GetDBWithContext(ctx).
			Where(filter).
			Order(clauseOrder).
			FindInBatches(&ptrs, batchSize, func(tx *gorm.DB, batch int) error {
				ctx = context.WithValue(ctx, contextTxKey{}, tx)
				return callback(ctx, batch, ptrs)
			})
		if result.Error != nil {
			log.Printf("find in batches by map filter (order) failed. table: %s, error: %v", result.Statement.Table, result.Error)
			return errorx.New(
				gormx.ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindInBatchesByMapFilter(Order)[%s]", result.Statement.Table),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find in batches by map filter (order) failed. table: %s, %s", result.Statement.Table, gormx.WarnNoRowsAffected)
		}
	*/
	ptrs := make([]PT, 0, batchSize)
	err := g.Session.FindInBatchesByMapFilter(ctx, &ptrs, filter, batchSize,
		func(ctx context.Context, batch int, models any) error {
			typedModels, ok := models.([]PT)
			if !ok {
				return errorx.NewWithDetails(
					gormx.ErrInvalidTypeAssertion,
					"gormx",
					"FindInBatchesByMapFilter",
					fmt.Sprintf("unexpected type: %T", models),
					nil,
				)
			}
			return callback(ctx, batch, typedModels)
		},
	)
	if err != nil {
		return err
	}
	return nil
}

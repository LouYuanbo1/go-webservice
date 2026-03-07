package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
)

func (s *session) FindByIDs(ctx context.Context, dest any, ids any, opts ...OrderOption) error {
	if ids == nil {
		log.Printf("find by ids failed : %s", WarnEmptyIDSlice)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = s.GetDBWithContext(ctx).
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

	result = s.GetDBWithContext(ctx).
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

func (s *session) FindByStructFilter(ctx context.Context, dest any, filter any, opts ...OrderOption) error {
	if filter == nil {
		log.Printf("find by struct filter failed : %s", WarnInvalidFilter)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = s.GetDBWithContext(ctx).
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

	result = s.GetDBWithContext(ctx).
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

func (s *session) FindByMapFilter(ctx context.Context, dest any, filter map[string]any, opts ...OrderOption) error {
	if len(filter) == 0 {
		log.Printf("find by map filter failed : %s", WarnInvalidFilter)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = s.GetDBWithContext(ctx).
			Where(filter).
			Find(dest)
		if result.Error != nil {
			log.Printf("find by map filter failed. table: %s, error: %v", result.Statement.Table, result.Error)
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindByMapFilter[%s]", result.Statement.Table),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by map filter failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
		}
		return nil
	}

	clauseOrder := s.clauseOrderBuilder(opts...)

	result = s.GetDBWithContext(ctx).
		Where(filter).
		Order(clauseOrder).
		Find(dest)
	if result.Error != nil {
		log.Printf("find by map filter failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindByMapFilter(Order)[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by map filter (order) failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (s *session) FindByPage(ctx context.Context, dest any, primaryKey string, page, pageSize int, opts ...OrderOption) error {
	if page <= 0 || pageSize <= 0 {
		log.Printf("find by page %d, pageSize %d failed : %s", page, pageSize, WarnInvalidPageParams)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = s.GetDBWithContext(ctx).
			Order(fmt.Sprintf("%s ASC", primaryKey)).
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

	result = s.GetDBWithContext(ctx).
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

func (s *session) FindByCursor(ctx context.Context, dest any, primaryKey string, cursor any, limit int) error {
	if limit <= 0 {
		log.Printf("find by cursor failed : %s", WarnInvalidLimit)
		return nil
	}
	result := s.GetDBWithContext(ctx).
		Where(fmt.Sprintf("%s > ?", primaryKey), cursor).
		Order(fmt.Sprintf("%s ASC", primaryKey)).
		Limit(limit + 1).
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

func (s *session) FindInBatches(
	ctx context.Context,
	dest any,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
	opts ...OrderOption,
) error {
	if batchSize <= 0 {
		log.Printf("find in batches failed : %s", WarnInvalidBatchSize)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = s.GetDBWithContext(ctx).
			FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
				//ctx = context.WithValue(ctx, contextTxKey{}, tx)
				return callback(ctx, tx, batch, dest)
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
			//ctx = context.WithValue(ctx, contextTxKey{}, tx)
			return callback(ctx, tx, batch, dest)
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

func (s *session) FindInBatchesByStructFilter(
	ctx context.Context,
	dest any,
	filter any,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
	opts ...OrderOption,
) error {
	if filter == nil {
		log.Printf("find in batches by struct filter failed : %s", WarnInvalidFilter)
		return nil
	}
	if batchSize <= 0 {
		log.Printf("find in batches by struct filter failed : %s", WarnInvalidBatchSize)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = s.GetDBWithContext(ctx).
			Where(filter).
			FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
				//ctx = context.WithValue(ctx, contextTxKey{}, tx)
				return callback(ctx, tx, batch, dest)
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

	result = s.GetDBWithContext(ctx).
		Where(filter).
		Order(clauseOrder).
		FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
			//ctx = context.WithValue(ctx, contextTxKey{}, tx)
			return callback(ctx, tx, batch, dest)
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

func (s *session) FindInBatchesByMapFilter(
	ctx context.Context,
	dest any,
	filter map[string]any,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
	opts ...OrderOption,
) error {
	if len(filter) == 0 {
		log.Printf("find in batches by map filter failed : %s", WarnInvalidFilter)
		return nil
	}
	if batchSize <= 0 {
		log.Printf("find in batches by map filter failed : %s", WarnInvalidBatchSize)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = s.GetDBWithContext(ctx).
			Where(filter).
			FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
				//ctx = context.WithValue(ctx, contextTxKey{}, tx)
				return callback(ctx, tx, batch, dest)
			})
		if result.Error != nil {
			log.Printf("find in batches by map filter failed. table: %s, error: %v", result.Statement.Table, result.Error)
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				fmt.Sprintf("FindInBatchesByMapFilter[%s]", result.Statement.Table),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find in batches by map filter failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
		}
		return nil
	}

	clauseOrder := s.clauseOrderBuilder(opts...)

	result = s.GetDBWithContext(ctx).
		Where(filter).
		Order(clauseOrder).
		FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
			//ctx = context.WithValue(ctx, contextTxKey{}, tx)
			return callback(ctx, tx, batch, dest)
		})
	if result.Error != nil {
		log.Printf("find in batches by map filter (order) failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("FindInBatchesByMapFilter(Order)[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find in batches by map filter (order) failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

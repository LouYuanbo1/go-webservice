package internal

import (
	"context"
	"log"

	"github.com/LouYuanbo1/go-webservice/gormx/errors"
	"github.com/LouYuanbo1/go-webservice/gormx/options"
	"gorm.io/gorm"
)

func (gx *gormX[T, ID, PT]) Create(ctx context.Context, model PT, opts ...options.ConflictOption) error {
	if model == nil {
		log.Printf("create failed : %s", errors.WarnInvalidModel)
		return nil
	}

	tableName := model.TableName()
	var result *gorm.DB
	// 应用冲突选项
	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			Create(model)
		if result.Error != nil {
			return errors.New(
				errors.ErrCreateFailed,
				"Create",
				tableName,
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("create failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
		}
		return nil
	}

	clauseConflict, err := gx.clauseOnConflictBuilder(opts...)
	if err != nil {
		return errors.New(
			errors.ErrInvalidOnConflictClause,
			"Create",
			tableName,
			err,
		)
	}

	result = gx.GetDBWithContext(ctx).
		Clauses(clauseConflict).
		Create(model)
	if result.Error != nil {
		return errors.New(
			errors.ErrCreateFailed,
			"Create(Upsert)",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("create(upsert) failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) CreateInBatches(ctx context.Context, models []PT, batchSize int, opts ...options.ConflictOption) error {
	// 参数校验
	if batchSize <= 0 {
		log.Printf("create in batches failed : %s", errors.WarnInvalidBatchSize)
		return nil
	}
	if len(models) == 0 {
		// 空切片属于合法操作（0 行插入），静默成功更符合批量操作语义
		log.Printf("skipped create in batches: %s", errors.WarnEmptyModelsSlice)
		return nil
	}

	tableName := models[0].TableName()
	var result *gorm.DB

	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			CreateInBatches(models, batchSize)
		if result.Error != nil {
			log.Printf("create in batches failed. table: %s, error: %v", tableName, result.Error)
			return errors.New(
				errors.ErrCreateFailed,
				"CreateInBatches",
				tableName,
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("create in batches failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
		}
		return nil
	}

	// 应用冲突选项
	clauseConflict, err := gx.clauseOnConflictBuilder(opts...)
	if err != nil {
		return errors.New(
			errors.ErrInvalidOnConflictClause,
			"CreateInBatches",
			tableName,
			err,
		)
	}

	result = gx.GetDBWithContext(ctx).
		Clauses(clauseConflict).
		CreateInBatches(models, batchSize)
	if result.Error != nil {
		log.Printf("create(upsert) in batches failed. table: %s, error: %v", tableName, result.Error)
		return errors.New(
			errors.ErrCreateFailed,
			"CreateInBatches(Upsert)",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("create in batches failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return nil
}

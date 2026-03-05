package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
)

func (gx *gormX[T, ID, PT]) Create(ctx context.Context, model PT, opts ...ConflictOption) error {
	if model == nil {
		log.Printf("create failed : %s", WarnInvalidModel)
		return nil
	}

	tableName := model.TableName()
	var result *gorm.DB
	// 应用冲突选项
	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			Create(model)
		if result.Error != nil {
			return errorx.New(
				ErrCreateFailed,
				"gormx",
				fmt.Sprintf("Create[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("create failed. table: %s, %s", tableName, WarnNoRowsAffected)
		}
		return nil
	}

	clauseConflict, err := gx.clauseOnConflictBuilder(opts...)
	if err != nil {
		return errorx.New(
			ErrInvalidOnConflictClause,
			"gormx",
			fmt.Sprintf("Create[%s]", tableName),
			err,
		)
	}

	result = gx.GetDBWithContext(ctx).
		Clauses(clauseConflict).
		Create(model)
	if result.Error != nil {
		return errorx.New(
			ErrCreateFailed,
			"gormx",
			fmt.Sprintf("Create(Upsert)[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("create(upsert) failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) CreateInBatches(ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) error {
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

	tableName := models[0].TableName()
	var result *gorm.DB

	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			CreateInBatches(models, batchSize)
		if result.Error != nil {
			log.Printf("create in batches failed. table: %s, error: %v", tableName, result.Error)
			return errorx.New(
				ErrCreateFailed,
				"gormx",
				fmt.Sprintf("CreateInBatches[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("create in batches failed. table: %s, %s", tableName, WarnNoRowsAffected)
		}
		return nil
	}

	// 应用冲突选项
	clauseConflict, err := gx.clauseOnConflictBuilder(opts...)
	if err != nil {
		return errorx.New(
			ErrInvalidOnConflictClause,
			"gormx",
			fmt.Sprintf("CreateInBatches[%s]", tableName),
			err,
		)
	}

	result = gx.GetDBWithContext(ctx).
		Clauses(clauseConflict).
		CreateInBatches(models, batchSize)
	if result.Error != nil {
		log.Printf("create(upsert) in batches failed. table: %s, error: %v", tableName, result.Error)
		return errorx.New(
			ErrCreateFailed,
			"gormx",
			fmt.Sprintf("CreateInBatches(Upsert)[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("create in batches failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return nil
}

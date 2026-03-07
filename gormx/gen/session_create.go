package gen

import (
	"context"
	"log"

	"github.com/LouYuanbo1/go-webservice/gormx"
)

func (g *genSession[T, ID, PT]) Create(ctx context.Context, model PT, opts ...gormx.ConflictOption) error {
	/*
		if model == nil {
			log.Printf("create failed : %s", gormx.WarnInvalidModel)
			return nil
		}

		var result *gorm.DB
		// 应用冲突选项
		if len(opts) == 0 {
			result = s.GetDBWithContext(ctx).
				Create(model)
			if result.Error != nil {
				return errorx.New(
					gormx.ErrCreateFailed,
					"gormx",
					"Create",
					result.Error,
				)
			}
			if result.RowsAffected == 0 {
				log.Printf("create failed. table: %s, %s", result.Statement.Table, gormx.WarnNoRowsAffected)
			}
			return nil
		}

		clauseConflict, err := gormx.ClauseOnConflictBuilder(opts...)
		if err != nil {
			return errorx.New(
				gormx.ErrInvalidOnConflictClause,
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
				gormx.ErrCreateFailed,
				"gormx",
				fmt.Sprintf("Create(Upsert)[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("create(upsert) failed. table: %s, %s", tableName, gormx.WarnNoRowsAffected)
		}
	*/
	return g.Session.Create(ctx, model, opts...)
}

func (g *genSession[T, ID, PT]) CreateInBatches(ctx context.Context, models []PT, batchSize int, opts ...gormx.ConflictOption) error {
	/*
		// 参数校验
		if batchSize <= 0 {
			log.Printf("create in batches failed : %s", gormx.WarnInvalidBatchSize)
			return nil
		}
		if len(models) == 0 {
			// 空切片属于合法操作（0 行插入），静默成功更符合批量操作语义
			log.Printf("skipped create in batches: %s", gormx.WarnEmptyModelsSlice)
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
					gormx.ErrCreateFailed,
					"gormx",
					fmt.Sprintf("CreateInBatches[%s]", tableName),
					result.Error,
				)
			}
			if result.RowsAffected == 0 {
				log.Printf("create in batches failed. table: %s, %s", tableName, gormx.WarnNoRowsAffected)
			}
			return nil
		}

		// 应用冲突选项
		clauseConflict, err := gormx.ClauseOnConflictBuilder(opts...)
		if err != nil {
			return errorx.New(
				gormx.ErrInvalidOnConflictClause,
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
				gormx.ErrCreateFailed,
				"gormx",
				fmt.Sprintf("CreateInBatches(Upsert)[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("create in batches failed. table: %s, %s", tableName, gormx.WarnNoRowsAffected)
		}
	*/
	if len(models) == 0 {
		// 空切片属于合法操作（0 行插入），静默成功更符合批量操作语义
		log.Printf("skipped create in batches: %s", gormx.WarnEmptyModelsSlice)
		return nil
	}
	return g.Session.CreateInBatches(ctx, models, batchSize, opts...)
}

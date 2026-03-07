package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
)

func (s *session) Create(ctx context.Context, model any, opts ...ConflictOption) error {
	if model == nil {
		log.Printf("create failed : %s", WarnInvalidModel)
		return nil
	}

	var result *gorm.DB
	// 应用冲突选项
	if len(opts) == 0 {
		result = s.GetDBWithContext(ctx).
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

	result = s.GetDBWithContext(ctx).
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
func (s *session) CreateInBatches(ctx context.Context, models any, batchSize int, opts ...ConflictOption) error {
	// 参数校验
	if batchSize <= 0 {
		log.Printf("create in batches failed : %s", WarnInvalidBatchSize)
		return nil
	}
	if models == nil {
		// 空切片属于合法操作（0 行插入），静默成功更符合批量操作语义
		log.Printf("skipped create in batches: %s", WarnEmptyModelsSlice)
		return nil
	}

	var result *gorm.DB

	if len(opts) == 0 {
		result = s.GetDBWithContext(ctx).
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

	result = s.GetDBWithContext(ctx).
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

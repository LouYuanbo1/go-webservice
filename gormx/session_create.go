package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
)

func (s *session) Create(ctx context.Context, model any, opts ...ConflictOption) error {
	prefix := "Create"
	if model == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidModel)
		return nil
	}

	// 1. 构建基础 DB 对象
	db := s.GetDBWithContext(ctx)

	// 2. 有冲突选项时，添加 ON CONFLICT 子句
	if len(opts) > 0 {
		clauseConflict, err := s.clauseOnConflictBuilder(opts...)
		if err != nil {
			return errorx.New(
				ErrInvalidOnConflictClause,
				"gormx",
				"Create",
				err,
			)
		}
		db = db.Clauses(clauseConflict)
		prefix = "Create(Upsert)"
	}

	// 3. 统一执行 Create
	result := db.Create(model)
	if result.Error != nil {
		return errorx.New(
			ErrCreateFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}

	// 4. 统一处理 RowsAffected 日志
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

// 注意这里的models需要在调用时传入一个slice类型的参数
func (s *session) CreateInBatches(ctx context.Context, models any, batchSize int, opts ...ConflictOption) error {
	prefix := "CreateInBatches"
	// 参数校验
	if batchSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidBatchSize)
		return nil
	}
	if models == nil {
		// 空切片属于合法操作（0 行插入），静默成功更符合批量操作语义
		log.Printf("%s skipped: %s", prefix, WarnEmptyModelsSlice)
		return nil
	}

	// 1. 构建基础 DB 对象
	db := s.GetDBWithContext(ctx)

	// 2. 有冲突选项时，添加 ON CONFLICT 子句
	if len(opts) > 0 {
		clauseConflict, err := s.clauseOnConflictBuilder(opts...)
		if err != nil {
			return errorx.New(
				ErrInvalidOnConflictClause,
				"gormx",
				"Create",
				err,
			)
		}
		db = db.Clauses(clauseConflict)
		prefix = "CreateInBatches(Upsert)"
	}

	result := db.CreateInBatches(models, batchSize)
	if result.Error != nil {
		return errorx.New(
			ErrCreateFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

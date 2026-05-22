package gormx

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
)

func (s *session) FindByIDs(ctx context.Context, dest any, ids any, opts ...OrderOption) error {
	prefix := "FindByIDs"
	if ids == nil {
		log.Printf("%s failed : %s", prefix, WarnEmptyIDSlice)
		return nil
	}

	// 1. 构建基础 DB 对象
	db := s.GetDBWithContext(ctx)

	// 2. 有排序选项时，添加 ORDER BY 子句
	if len(opts) > 0 {
		clauseOrder := s.clauseOrderBuilder(opts...)
		db = db.Order(clauseOrder)
		prefix = "FindByIDs(Order)"
	}

	result := db.Find(dest, ids)
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
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

func (s *session) FindByStructFilter(ctx context.Context, dest any, filter any, opts ...OrderOption) error {
	prefix := "FindByStructFilter"
	if filter == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidFilter)
		return nil
	}

	// 1. 构建基础 DB 对象
	db := s.GetDBWithContext(ctx)

	// 2. 有排序选项时，添加 ORDER BY 子句
	if len(opts) > 0 {
		clauseOrder := s.clauseOrderBuilder(opts...)
		db = db.Order(clauseOrder)
		prefix = "FindByStructFilter(Order)"
	}

	result := db.Where(filter).Find(dest)
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
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

func getPrimaryKeyColumns(db *gorm.DB, dest any) (string, error) {
	// 独立会话，DryRun 防止意外执行 SQL
	stmt := db.Session(&gorm.Session{DryRun: true}).Model(dest).Statement
	// 显式解析 dest，填充 Schema
	if err := stmt.Parse(stmt.Model); err != nil {
		return "", fmt.Errorf("解析模型失败: %w", err)
	}
	if stmt.Schema == nil {
		return "", fmt.Errorf("无法解析模型 Schema")
	}
	fields := stmt.Schema.PrimaryFieldDBNames
	if len(fields) == 0 {
		return "", fmt.Errorf("模型未定义主键")
	}
	return strings.Join(fields, ", "), nil
}

func (s *session) FindByPage(ctx context.Context, dest any, page, pageSize int, opts ...OrderOption) error {
	prefix := "FindByPage"
	if page <= 0 || pageSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidPageParams)
		return nil
	}

	// 1. 构建基础 DB 对象
	db := s.GetDBWithContext(ctx).Model(dest)

	// 2. 有排序选项时，添加 ORDER BY 子句
	if len(opts) > 0 {
		clauseOrder := s.clauseOrderBuilder(opts...)
		db = db.Order(clauseOrder)
		prefix = "FindByPage(Order)"
	} else {
		// 自动获取主键排序
		pkColumns, err := getPrimaryKeyColumns(db, dest)
		if err != nil {
			return errorx.New(
				ErrQueryFailed,
				"gormx",
				"FindByPage: get primary key",
				err,
			)
		}
		clauseOrder := fmt.Sprintf("%s ASC", pkColumns)
		db = db.Order(clauseOrder)
	}

	result := db.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(dest)

	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
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

func (s *session) FindByCursor(ctx context.Context, dest any, cursor any, limit int) error {
	prefix := "FindByCursor"
	if limit <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidLimit)
		return nil
	}

	db := s.GetDBWithContext(ctx).Model(dest)
	pkColumns, err := getPrimaryKeyColumns(db, dest)
	if err != nil {
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("%s: get primary key", prefix),
			err,
		)
	}
	result := db.
		Where(fmt.Sprintf("%s > ?", pkColumns), cursor).
		Order(fmt.Sprintf("%s ASC", pkColumns)).
		Limit(limit).
		Find(dest)
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
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

func (s *session) FindInBatches(
	ctx context.Context,
	dest any,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
	opts ...OrderOption,
) error {
	prefix := "FindInBatches"
	if batchSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidBatchSize)
		return nil
	}

	// 1. 构建基础 DB 对象
	db := s.GetDBWithContext(ctx)

	// 2. 有排序选项时，添加 ORDER BY 子句
	if len(opts) > 0 {
		clauseOrder := s.clauseOrderBuilder(opts...)
		db = db.Order(clauseOrder)
		prefix = "FindInBatches(Order)"
	}

	result := db.FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
		return callback(ctx, tx, batch, dest)
	})
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
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

func (s *session) FindInBatchesByStructFilter(
	ctx context.Context,
	dest any,
	filter any,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
	opts ...OrderOption,
) error {
	prefix := "FindInBatchesByStructFilter"
	if filter == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidFilter)
		return nil
	}
	if batchSize <= 0 {
		log.Printf("%s failed : %s", prefix, WarnInvalidBatchSize)
		return nil
	}

	// 1. 构建基础 DB 对象
	db := s.GetDBWithContext(ctx)

	// 2. 有排序选项时，添加 ORDER BY 子句
	if len(opts) > 0 {
		clauseOrder := s.clauseOrderBuilder(opts...)
		db = db.Order(clauseOrder)
		prefix = "FindInBatchesByStructFilter(Order)"
	}

	result := db.Where(filter).
		FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
			return callback(ctx, tx, batch, dest)
		})
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
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

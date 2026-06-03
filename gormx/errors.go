package gormx

import (
	"errors"
)

// 预定义的警告信息（用于业务层提示，非错误）
var (
	WarnInvalidModel      = "gormx: invalid model"
	WarnEmptyModelsSlice  = "gormx: empty models slice"
	WarnInvalidBatchSize  = "gormx: invalid batch size"
	WarnInvalidID         = "gormx: invalid id"
	WarnEmptyIDSlice      = "gormx: empty id slice"
	WarnInvalidFilter     = "gormx: invalid filter"
	WarnInvalidPageParams = "gormx: invalid page parameters"
	WarnInvalidLimit      = "gormx: invalid limit"
	WarnInvalidUpdateData = "gormx: invalid update data"
	WarnNoRowsAffected    = "gormx: no rows affected"
)

// 预定义的 Sentinel 错误（用于 errors.Is 判断）
var (
	// 数据库连接错误
	ErrInvalidInitConfig  = errors.New("gormx: invalid init config")
	ErrDBConnection       = errors.New("gormx: database connection error")
	ErrExecutionSQLScript = errors.New("gormx: execution sql script error")
	// 冲突处理错误
	ErrInvalidConflictStrategy = errors.New("gormx: invalid conflict strategy")
	ErrEmptyUpdateColumns      = errors.New("gormx: empty update columns")
	// 约束错误
	ErrEmptyConstraint         = errors.New("gormx: empty constraint")
	ErrInvalidOnConflictClause = errors.New("gormx: invalid on conflict clause")
	// 数据库操作错误
	ErrCreateFailed      = errors.New("gormx: create failed")
	ErrQueryFailed       = errors.New("gormx: query failed")
	ErrUpdateFailed      = errors.New("gormx: update failed")
	ErrDeleteFailed      = errors.New("gormx: delete failed")
	ErrTransactionFailed = errors.New("gormx: transaction failed")
	// 解析模型失败
	ErrParseModelFailed = errors.New("gormx: parse model failed")
	// 模型未定义主键
	ErrModelNoPrimaryKey = errors.New("gormx: model no primary key")
		
	//断言错误
	ErrInvalidTypeAssertion = errors.New("gormx: invalid type assertion")
)

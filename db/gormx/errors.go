package gormx

import (
	"errors"
)

// 预定义的 Sentinel 错误（用于 errors.Is 判断）
var (
	// 数据库连接错误
	ErrInvalidInitConfig  = errors.New("gormx: invalid init config")
	ErrDBConnection       = errors.New("gormx: database connection error")
	ErrExecutionSQLScript = errors.New("gormx: execution sql script error")
	//模型传入错误/过滤条件错误
	ErrInvalidModel  = errors.New("invalid model")
	ErrInvalidFilter = errors.New("invalid filter")
	// 数据库操作错误
	ErrNoRowsAffected = errors.New("no rows affected")
	ErrExecFailed     = errors.New("exec failed")
	ErrCreateFailed   = errors.New("create failed")
	ErrFirstFailed    = errors.New("first failed")
	ErrPluckFailed    = errors.New("pluck failed")
	ErrScanFailed     = errors.New("scan failed")
	ErrFindFailed     = errors.New("find failed")
	ErrCountFailed    = errors.New("count failed")
	ErrUpdateFailed   = errors.New("update failed")
	ErrUpdatesFailed  = errors.New("updates failed")
	ErrDeleteFailed   = errors.New("delete failed")
)

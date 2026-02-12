package errors

import (
	"errors"
	"fmt"
)

var (
	// 参数验证错误
	WarnInvalidModel      = "gormx: invalid model"
	WarnEmptyModelsSlice  = "gormx: empty models slice"
	WarnInvalidBatchSize  = "gormx: invalid batch size"
	WarnInvalidID         = "gormx: invalid id"
	WarnEmptyIDsSlice     = "gormx: empty ids slice"
	WarnInvalidFilter     = "gormx: invalid filter"
	WarnInvalidPageParams = "gormx: invalid page parameters"
	WarnInvalidLimit      = "gormx: invalid limit"
	WarnInvalidUpdateData = "gormx: invalid update data"
	WarnNoRowsAffected    = "gormx: no rows affected"
)

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
	ErrCreateFailed = errors.New("gormx: create failed")
	ErrQueryFailed  = errors.New("gormx: query failed")
	ErrUpdateFailed = errors.New("gormx: update failed")
	ErrDeleteFailed = errors.New("gormx: delete failed")
)

// 带上下文的错误类型
type Error struct {
	// 错误类型 Error Type
	Err error
	// 操作名称 Operation
	Op string
	// 表名 TableName
	Table string
	// 详细信息 Details
	Details string
	// 原始错误 Original error (e.g. gorm.Error)
	Cause error
}

func (e *Error) Error() string {
	if e.Table != "" && e.Op != "" {
		return fmt.Sprintf("gormx.%s[%s]: %s: %v", e.Op, e.Table, e.Err, e.Cause)
	}
	if e.Cause != nil {
		return fmt.Sprintf("gormx: %v: %v", e.Err, e.Cause)
	}
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

// 错误构建函数
func New(err error, op, table string, cause error) error {
	return &Error{
		Err:   err,
		Op:    op,
		Table: table,
		Cause: cause,
	}
}

func NewWithDetails(err error, op, table, details string, cause error) error {
	return &Error{
		Err:     err,
		Op:      op,
		Table:   table,
		Details: details,
		Cause:   cause,
	}
}

func IsInvalidInitConfig(err error) bool {
	return errors.Is(err, ErrInvalidInitConfig)
}

func IsDBConnection(err error) bool {
	return errors.Is(err, ErrDBConnection)
}

func IsExecutionSQLScript(err error) bool {
	return errors.Is(err, ErrExecutionSQLScript)
}

func IsInvalidConflictStrategy(err error) bool {
	return errors.Is(err, ErrInvalidConflictStrategy)
}

func IsEmptyUpdateColumns(err error) bool {
	return errors.Is(err, ErrEmptyUpdateColumns)
}

func IsEmptyConstraint(err error) bool {
	return errors.Is(err, ErrEmptyConstraint)
}
func IsInvalidOnConflictClause(err error) bool {
	return errors.Is(err, ErrInvalidOnConflictClause)
}

func IsCreateFailed(err error) bool {
	return errors.Is(err, ErrCreateFailed)
}

func IsQueryFailed(err error) bool {
	return errors.Is(err, ErrQueryFailed)
}

func IsUpdateFailed(err error) bool {
	return errors.Is(err, ErrUpdateFailed)
}

func IsDeleteFailed(err error) bool {
	return errors.Is(err, ErrDeleteFailed)
}

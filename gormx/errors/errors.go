package errors

import (
	"errors"
	"fmt"
	"strings"
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
	ErrCreateFailed = errors.New("gormx: create failed")
	ErrQueryFailed  = errors.New("gormx: query failed")
	ErrUpdateFailed = errors.New("gormx: update failed")
	ErrDeleteFailed = errors.New("gormx: delete failed")
)

// Error 是 gormx 自定义的错误类型，包装了语义错误、操作上下文和原始底层错误。
type Error struct {
	// Err 是当前层定义的语义错误，通常为上述 Sentinel 错误之一
	Err error
	// Op 是发生错误时的操作名称，例如 "Create", "Update"
	Op string
	// Table 是相关的数据库表名
	Table string
	// Details 是额外的上下文信息，例如具体的字段值或 SQL 片段
	Details string
	// Cause 是原始的底层错误，例如来自 GORM 的 error
	Cause error
}

// Error 实现 error 接口，生成格式化的错误消息。
// 格式为：gormx[.Op][.Table]: Err | details=... | cause=...
func (e *Error) Error() string {
	// 1. 防御性检查：避免 e.Err 为 nil 时 Panic
	if e.Err == nil {
		if e.Cause != nil {
			return fmt.Sprintf("gormx: unknown error (caused by: %v)", e.Cause)
		}
		return "gormx: unknown error"
	}

	// 2. 构建动态前缀，尽量保留可用的上下文信息
	var prefix string
	switch {
	case e.Op != "" && e.Table != "":
		prefix = fmt.Sprintf("gormx.%s[%s]", e.Op, e.Table)
	case e.Op != "":
		prefix = fmt.Sprintf("gormx.%s", e.Op)
	case e.Table != "":
		prefix = fmt.Sprintf("gormx[%s]", e.Table)
	default:
		prefix = "gormx"
	}

	// 3. 使用 strings.Builder 高效拼接
	var buf strings.Builder
	buf.Grow(len(prefix) + len(e.Err.Error()) + 64) // 预分配足够空间

	buf.WriteString(prefix)
	buf.WriteString(": ")
	buf.WriteString(e.Err.Error())

	// 4. 附加 Details（如果有）
	if e.Details != "" {
		buf.WriteString(" | details=")
		buf.WriteString(e.Details)
	}

	// 5. 附加 Cause（如果有）
	if e.Cause != nil {
		buf.WriteString(" | cause=")
		buf.WriteString(e.Cause.Error())
	}

	return buf.String()
}

// Unwrap 返回底层的原始错误（Cause），使错误链能够遍历到 GORM 等原始错误。
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is 使 errors.Is 能够匹配到 Error 中包装的语义错误（Err）。
// 例如 errors.Is(err, ErrCreateFailed) 将返回 true。
func (e *Error) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// New 创建一个新的 Error，包含语义错误、操作名、表名和底层原因。
func New(err error, op, table string, cause error) error {
	return &Error{
		Err:   err,
		Op:    op,
		Table: table,
		Cause: cause,
	}
}

// NewWithDetails 创建一个带 Details 字段的 Error。
func NewWithDetails(err error, op, table, details string, cause error) error {
	return &Error{
		Err:     err,
		Op:      op,
		Table:   table,
		Details: details,
		Cause:   cause,
	}
}

// 以下为辅助函数，方便调用者直接检查特定 Sentinel 错误。
// 它们内部使用 errors.Is，可以正确处理任何包装了 Sentinel 的错误（包括 *Error）。

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

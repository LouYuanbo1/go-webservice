package errors

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrCheckExistence = errors.New("elasticsearchx: check index existence failed")
	ErrGetMapping     = errors.New("elasticsearchx: get index mapping failed")
	ErrMarshalMapping = errors.New("elasticsearchx: marshal index mapping failed")
	ErrCreateIndex    = errors.New("elasticsearchx: create index failed")
	ErrDeleteIndex    = errors.New("elasticsearchx: delete index failed")
)

// Error 是 gormx 自定义的错误类型，包装了语义错误、操作上下文和原始底层错误。
type Error struct {
	// Err 是当前层定义的语义错误，通常为上述 Sentinel 错误之一
	Err error
	// Op 是发生错误时的操作名称，例如 "Create", "Update"
	Op string
	// Table 是相关的数据库表名
	Index string
	// Details 是额外的上下文信息，例如具体的字段值或 SQL 片段
	Details string
	// Cause 是原始的底层错误，例如来自 GORM 的 error
	Cause error
}

// Error 实现 error 接口，生成格式化的错误消息。
// 格式为：elasticsearchx[.Op][.Index]: Err | details=... | cause=...
func (e *Error) Error() string {
	// 1. 防御性检查：避免 e.Err 为 nil 时 Panic
	if e.Err == nil {
		if e.Cause != nil {
			return fmt.Sprintf("elasticsearchx: unknown error (caused by: %v)", e.Cause)
		}
		return "elasticsearchx: unknown error"
	}

	// 2. 构建动态前缀，尽量保留可用的上下文信息
	var prefix string
	switch {
	case e.Op != "" && e.Index != "":
		prefix = fmt.Sprintf("elasticsearchx.%s[%s]", e.Op, e.Index)
	case e.Op != "":
		prefix = fmt.Sprintf("elasticsearchx.%s", e.Op)
	case e.Index != "":
		prefix = fmt.Sprintf("elasticsearchx[%s]", e.Index)
	default:
		prefix = "elasticsearchx"
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
func New(err error, op, index string, cause error) error {
	return &Error{
		Err:   err,
		Op:    op,
		Index: index,
		Cause: cause,
	}
}

// NewWithDetails 创建一个带 Details 字段的 Error。
func NewWithDetails(err error, op, index, details string, cause error) error {
	return &Error{
		Err:     err,
		Op:      op,
		Index:   index,
		Details: details,
		Cause:   cause,
	}
}

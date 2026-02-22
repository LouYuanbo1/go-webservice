package errors

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrLoadImage   = errors.New("imgutil: load image error")
	ErrSaveImage   = errors.New("imgutil: save image error")
	ErrDeleteImage = errors.New("imgutil: delete image error")
)

// 带上下文的错误类型
type Error struct {
	// 错误类型 Error Type
	Err error
	// 操作名称 Operation
	Op string
	// 详细信息 Details
	Details string
	// 原始错误 Original error (e.g. gorm.Error)
	Cause error
}

func (e *Error) Error() string {
	// 1. 防御性检查：避免 e.Err 为 nil 时 Panic
	if e.Err == nil {
		if e.Cause != nil {
			return fmt.Sprintf("imgutil: unknown error (caused by: %v)", e.Cause)
		}
		return "imgutil: unknown error"
	}

	// 2. 构建动态前缀，尽量保留可用的上下文信息
	var prefix string
	switch {
	case e.Op != "":
		prefix = fmt.Sprintf("imgutil.%s", e.Op)
	default:
		prefix = "imgutil"
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

func (e *Error) Unwrap() error {
	return e.Cause
}

func (e *Error) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// 错误构建函数
func New(err error, op string, cause error) error {
	return &Error{
		Err:   err,
		Op:    op,
		Cause: cause,
	}
}

func NewWithDetails(err error, op string, details string, cause error) error {
	return &Error{
		Err:     err,
		Op:      op,
		Details: details,
		Cause:   cause,
	}
}

func IsLoadImage(err error) bool {
	return errors.Is(err, ErrLoadImage)
}

func IsSaveImage(err error) bool {
	return errors.Is(err, ErrSaveImage)
}

func IsDeleteImage(err error) bool {
	return errors.Is(err, ErrDeleteImage)
}

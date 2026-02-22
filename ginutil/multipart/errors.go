package multipart

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidObject         = errors.New("ginutil/multipart: invalid object")
	ErrParseMultipartForm    = errors.New("ginutil/multipart: parse multipart form error")
	ErrDecodeForm            = errors.New("ginutil/multipart: decode form error")
	ErrFillFiles             = errors.New("ginutil/multipart: fill files error")
	ErrDuplicateBracketIndex = errors.New("ginutil/multipart: duplicate bracket index")
	ErrDuplicateDotIndex     = errors.New("ginutil/multipart: duplicate dot index")
	ErrInvalidIndex          = errors.New("ginutil/multipart: invalid index")
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
			return fmt.Sprintf("ginutil/multipart: unknown error (caused by: %v)", e.Cause)
		}
		return "ginutil/multipart: unknown error"
	}

	// 2. 构建动态前缀，尽量保留可用的上下文信息
	var prefix string
	switch {
	case e.Op != "":
		prefix = fmt.Sprintf("ginutil/multipart.%s", e.Op)
	default:
		prefix = "ginutil/multipart"
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

func NewWithDetails(err error, op, details string, cause error) error {
	return &Error{
		Err:     err,
		Op:      op,
		Details: details,
		Cause:   cause,
	}
}

func IsInvalidObject(err error) bool {
	return errors.Is(err, ErrInvalidObject)
}

func IsParseMultipartForm(err error) bool {
	return errors.Is(err, ErrParseMultipartForm)
}

func IsDecodeForm(err error) bool {
	return errors.Is(err, ErrDecodeForm)
}

func IsFillFiles(err error) bool {
	return errors.Is(err, ErrFillFiles)
}

func IsDuplicateBracketIndex(err error) bool {
	return errors.Is(err, ErrDuplicateBracketIndex)
}

func IsDuplicateDotIndex(err error) bool {
	return errors.Is(err, ErrDuplicateDotIndex)
}

func IsInvalidIndex(err error) bool {
	return errors.Is(err, ErrInvalidIndex)
}

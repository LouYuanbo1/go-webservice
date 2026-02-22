package errors

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrJsonMarshal = errors.New("redisx: json marshal error")
	ErrJsonUnmarshal = errors.New("redisx: json unmarshal error")
	ErrExpire = errors.New("redisx: expire error")
	ErrSet  = errors.New("redisx: set error")
	ErrHSet = errors.New("redisx: hset error")
	ErrGet  = errors.New("redisx: get error")
	ErrHGet = errors.New("redisx: hget error")
	ErrHMGet = errors.New("redisx: hmget error")
	ErrHGetAll = errors.New("redisx: hgetall error")
	ErrNewDecoder = errors.New("redisx: new decoder error")
	ErrDecode = errors.New("redisx: decode error")
	ErrDel = errors.New("redisx: del error")
	ErrAcquire = errors.New("redisx: acquire error")
	ErrRelease = errors.New("redisx: release error")
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
			return fmt.Sprintf("redisx: unknown error (caused by: %v)", e.Cause)
		}
		return "redisx: unknown error"
	}

	// 2. 构建动态前缀，尽量保留可用的上下文信息
	var prefix string
	switch {
	case e.Op != "":
		prefix = fmt.Sprintf("redisx.%s", e.Op)
	default:
		prefix = "redisx"
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

func IsJsonMarshal(err error) bool {
	return errors.Is(err, ErrJsonMarshal)
}

func IsJsonUnmarshal(err error) bool {
	return errors.Is(err, ErrJsonUnmarshal)
}

func IsExpire(err error) bool {
	return errors.Is(err, ErrExpire)
}

func IsSet(err error) bool {
	return errors.Is(err, ErrSet)
}

func IsHSet(err error) bool {
	return errors.Is(err, ErrHSet)
}

func IsGet(err error) bool {
	return errors.Is(err, ErrGet)
}

func IsHGet(err error) bool {
	return errors.Is(err, ErrHGet)
}

func IsHMGet(err error) bool {
	return errors.Is(err, ErrHMGet)
}

func IsHGetAll(err error) bool {
	return errors.Is(err, ErrHGetAll)
}

func IsNewDecoder(err error) bool {
	return errors.Is(err, ErrNewDecoder)
}

func IsDecode(err error) bool {
	return errors.Is(err, ErrDecode)
}

func IsDel(err error) bool {
	return errors.Is(err, ErrDel)
}

func IsAcquire(err error) bool {
	return errors.Is(err, ErrAcquire)
}

func IsRelease(err error) bool {
	return errors.Is(err, ErrRelease)
}

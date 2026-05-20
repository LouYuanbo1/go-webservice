package breaker

import "errors"

// ========== 错误定义 ==========

var (
	ErrServiceUnavailable = errors.New("breaker: service unavailable")
	ErrRejected           = errors.New("breaker: request rejected")
)

package redis

import "errors"

var (
	ErrConnectFailed   = errors.New("event/redis: connect failed")
	ErrSubscribeFailed = errors.New("event/redis: subscribe failed")
	ErrPublishFailed   = errors.New("event/redis: publish failed")
	ErrClosed          = errors.New("event/redis: bus closed")
)

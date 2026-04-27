package cache

import (
	"fmt"

	"github.com/LouYuanbo1/go-webservice/errorx"
)

type Client interface {
	Cache
}

func Open(driver Driver, opts ...Option) (Client, error) {
	if driver == nil {
		return nil, errorx.NewWithDetails(
			ErrInit,
			"cache",
			"Open",
			"driver cannot be nil",
			nil,
		)
	}
	// 调用驱动的 Initialize 方法，完成具体初始化
	// 这里可以统一加入日志、监控等中间件逻辑
	cache, err := driver.Initialize()
	if err != nil {
		return nil, errorx.NewWithDetails(
			ErrInit,
			"cache",
			"Open",
			"initialize driver failed",
			err,
		)
	}
	fmt.Printf("[Cache] Initialized driver: %s\n", driver.Name())
	switch driver.Name() {
	case "local":
		return newLocalClient(cache.(LocalCache), opts...), nil
	case "redis":
		return newRedisClient(cache.(RedisCache), opts...), nil
	default:
		return nil, errorx.NewWithDetails(
			ErrDriverNotFound,
			"cache",
			"Open",
			"driver not found",
			nil,
		)
	}
}

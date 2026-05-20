package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/LouYuanbo1/go-webservice/errorx"
)

type Client struct {
	cache Cache
}

func Open(driver Driver) (*Client, error) {
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
	return &Client{cache: cache}, nil
}

func (c *Client) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	return c.cache.Set(ctx, key, val, expiration)
}

func (c *Client) Get(ctx context.Context, key string, val any) error {
	return c.cache.Get(ctx, key, val)
}

func (c *Client) Take(ctx context.Context, key string, val any,
	query func(val any) error, ttl time.Duration) error {
	return c.cache.Take(ctx, key, val, query, ttl)
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.cache.Del(ctx, keys...)
}

func (c *Client) GetRawCache() Cache {
	return c.cache
}

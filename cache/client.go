package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/LouYuanbo1/go-webservice/errorx"
)

type Client interface {
	Cache
}

type client struct {
	cache Cache
	opts  *Options
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
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	fmt.Printf("[Cache] Initialized driver: %s\n", driver.Name())
	return &client{
		cache: cache,
		opts:  options,
	}, nil
}

func (c *client) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	return c.cache.Set(ctx, c.opts.Prefix+key, val, expiration)
}

func (c *client) Get(ctx context.Context, key string, val any) error {
	return c.cache.Get(ctx, c.opts.Prefix+key, val)
}

func (c *client) Take(ctx context.Context, val any, key string,
	query func(val any) error, ttl time.Duration) error {
	return c.cache.Take(ctx, val, c.opts.Prefix+key, query, ttl)
}

func (c *client) Del(ctx context.Context, keys ...string) error {
	for i, key := range keys {
		keys[i] = c.opts.Prefix + key
	}
	return c.cache.Del(ctx, keys...)
}

package cache

import (
	"context"
	"time"
)

type Client struct {
	Cache
}

func NewClient(cache Cache) *Client {
	return &Client{Cache: cache}
}

func (tc *Client) GetRawCache() Cache {
	return tc.Cache
}

func (tc *Client) Set[T any](ctx context.Context, key string, val T, ttl time.Duration) error {
	return tc.Cache.Set(ctx, key, val, ttl)
}

// 预计同时支持传入Model和*Model,以便使用者自己判断是否需要返回指针
func (tc *Client) Get[T any](ctx context.Context, key string, val *T) error {
	return tc.Cache.Get(ctx, key, val)
}

func (tc *Client) Take[T any](ctx context.Context, key string, val *T,
	query func(val *T) error, ttl time.Duration) error {
	return tc.Cache.Take(ctx, key, val, func(val any) error { return query(val.(*T)) }, ttl)
}

func (tc *Client) Del(ctx context.Context, keys ...string) error {
	return tc.Cache.Del(ctx, keys...)
}

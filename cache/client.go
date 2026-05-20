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

func (c *Client) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	return c.Cache.Set(ctx, key, val, expiration)
}

func (c *Client) Get(ctx context.Context, key string, val any) error {
	return c.Cache.Get(ctx, key, val)
}

func (c *Client) Take(ctx context.Context, key string, val any,
	query func(val any) error, ttl time.Duration) error {
	return c.Cache.Take(ctx, key, val, query, ttl)
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.Cache.Del(ctx, keys...)
}

func (c *Client) GetRawCache() Cache {
	return c.Cache
}

package cache

/*
import (
	"context"
	"time"
)

type Client127 struct {
	Cache
}

func NewClient127(cache Cache) *Client127 {
	return &Client127{Cache: cache}
}

func (c *Client127) Set[T any](ctx context.Context, key string, val *T, ttl time.Duration) error {
	return c.Cache.Set(ctx, key, val, ttl)
}

func (c *Client127) Get[T any](ctx context.Context, key string) (*T, error) {
	var val T
	err := c.Cache.Get(ctx, key, &val)
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func (c *Client127) Take[T any](ctx context.Context, key string,
		query func(val *T) error, ttl time.Duration) (*T, error) {
	var val T
	err := c.Cache.Take(ctx, key, &val, query, ttl)
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func (c *Client127) Del(ctx context.Context, keys ...string) error {
	return c.Cache.Del(ctx, keys...)
}
*/

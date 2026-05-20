package cache

/*
import (
	"context"
	"time"
)

type TypedClient struct {
	Cache
}

func NewTypedClient(cache Cache) *TypedClient {
	return &TypedClient{Cache: cache}
}

func (tc *TypedClient) Set[T any](ctx context.Context, key string, val T, ttl time.Duration) error {
	return tc.Cache.Set(ctx, key, val, ttl)
}

//预计同时支持传入Model和*Model,以便使用者自己判断是否需要返回指针
func (tc *TypedClient) Get[T any](ctx context.Context, key string) (T, error) {
	var val T
	err := tc.Cache.Get(ctx, key, &val)
	if err != nil {
		var zero T
		return zero, err
	}
	return val, nil
}

func (tc *TypedClient) Take[T any](ctx context.Context, key string,
		query func(val *T) error, ttl time.Duration) (T, error) {
	var val T
	err := tc.Cache.Take(ctx, key, &val, query, ttl)
	if err != nil {
		var zero T
		return zero, err
	}
	return val, nil
}

func (tc *TypedClient) Del(ctx context.Context, keys ...string) error {
	return tc.Cache.Del(ctx, keys...)
}
*/

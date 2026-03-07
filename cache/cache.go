package cache

import (
	"context"
	"time"
)

type Cache[T any] interface {
	// Del deletes cached values with keys.
	Del(ctx context.Context, keys ...string) error
	// Get gets the cache with key and fills into v.
	Get(ctx context.Context, key string, val any) error
	// IsNotFound checks if the given error is the defined errNotFound.
	IsNotFound(err error) bool
	// Set sets the cache with key and v, using c.expiry.
	Set(ctx context.Context, key string, val any) error
	// SetWithExpire sets the cache with key and v, using given expire.
	SetWithExpire(ctx context.Context, key string, val any, expire time.Duration) error
	// Take takes the result from cache first, if not found,
	// query from DB and set cache using c.expiry, then return the result.
	Take(ctx context.Context, val any, key string, query func(val any) error) error
	// TakeWithExpire takes the result from cache first, if not found,
	// query from DB and set cache using given expire, then return the result.
	TakeWithExpire(ctx context.Context, val any, key string,
		query func(val any, expire time.Duration) error) error
}

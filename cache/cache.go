package cache

import (
	"context"
	"time"
)

type Cache interface {
	// Set sets the cache with key and v, using c.expiry.
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	// Get gets the cache with key and fills into v.
	Get(ctx context.Context, key string, val any) error
	// Take takes the result from cache first, if not found,
	// query from DB and set cache using given expire, then return the result.
	Take(ctx context.Context, val any, key string,
		query func(val any) error, ttl time.Duration) error
	// Del deletes cached values with keys.
	Del(ctx context.Context, keys ...string) error
}

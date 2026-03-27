package cache

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

type LocalClient interface {
	LocalCache
}

type localClient struct {
	cache LocalCache
	opts  *Options
}

func newLocalClient(cache LocalCache, opts ...Option) LocalClient {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	return &localClient{cache: cache, opts: options}
}
func (lc *localClient) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	return lc.cache.Set(ctx, lc.opts.Prefix+key, val, expiration)
}

func (lc *localClient) Get(ctx context.Context, key string, val any) error {
	return lc.cache.Get(ctx, lc.opts.Prefix+key, val)
}

func (lc *localClient) Take(ctx context.Context, val any, key string,
	query func(val any) error, ttl time.Duration) error {
	return lc.cache.Take(ctx, val, lc.opts.Prefix+key, query, ttl)
}

func (lc *localClient) Del(ctx context.Context, keys ...string) error {
	for i, key := range keys {
		keys[i] = lc.opts.Prefix + key
	}
	return lc.cache.Del(ctx, keys...)
}

func (lc *localClient) GetLocalCache() *ristretto.Cache[string, any] {
	return lc.cache.GetLocalCache()
}

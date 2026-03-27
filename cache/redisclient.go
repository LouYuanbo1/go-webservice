package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	RedisCache
}

type redisClient struct {
	cache RedisCache
	opts  *Options
}

func NewRedisClient(cache RedisCache, opts ...Option) RedisClient {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	return &redisClient{cache: cache, opts: options}
}
func (rc *redisClient) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	return rc.cache.Set(ctx, rc.opts.Prefix+key, val, expiration)
}

func (rc *redisClient) Get(ctx context.Context, key string, val any) error {
	return rc.cache.Get(ctx, rc.opts.Prefix+key, val)
}

func (rc *redisClient) Take(ctx context.Context, val any, key string,
	query func(val any) error, ttl time.Duration) error {
	return rc.cache.Take(ctx, val, rc.opts.Prefix+key, query, ttl)
}

func (rc *redisClient) Del(ctx context.Context, keys ...string) error {
	for i, key := range keys {
		keys[i] = rc.opts.Prefix + key
	}
	return rc.cache.Del(ctx, keys...)
}

func (rc *redisClient) GetRedisClient() *redis.Client {
	return rc.cache.GetRedisClient()
}

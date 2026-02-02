package redisx

import (
	"context"
	"time"

	"github.com/LouYuanbo1/go-webservice/redisx/internal"
	"github.com/redis/go-redis/v9"
)

type RedisX[T any] interface {
	SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) error
	SetWithDefaultTTL(ctx context.Context, key string, value T) error
	HSetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) error
	HSetWithDefaultTTL(ctx context.Context, key string, value T) error
	Get(ctx context.Context, key string) (T, error)
	GetPointer(ctx context.Context, key string) (*T, error)
	HGet(ctx context.Context, key string, field string) (string, error)
	HMGet(ctx context.Context, key string, fields ...string) ([]any, error)
	HGetAll(ctx context.Context, key string) (T, error)
	HGetAllPointer(ctx context.Context, key string) (*T, error)
	Del(ctx context.Context, key string) error
	Acquire(ctx context.Context, key string, expire time.Duration) (string, bool, error)
	Release(ctx context.Context, key, lockID string) error
}

func NewRedisX[T any](client *redis.Client, defaultTTLKey time.Duration) RedisX[T] {
	return internal.NewRedisX[T](client, defaultTTLKey)
}

package localcache

import (
	"context"
	"fmt"
	"log"

	"github.com/dgraph-io/ristretto/v2"
)

type LocalCache[T any] interface {
	// Set sets the value for the given key.
	SetWithTTL(ctx context.Context, key string, value T, opts ...TTLOption) bool
	// Get gets the value for the given key.
	Get(ctx context.Context, key string) (T, bool)
	// GetPointer gets the pointer value for the given key.
	GetPointer(ctx context.Context, key string) (*T, bool)

	// Delete deletes the value for the given key.
	Del(ctx context.Context, key string)
}

type localCache[T any] struct {
	local  *ristretto.Cache[string, T]
	config OperationConfig
}

func NewLocalCache[T any](config *Config) (*localCache[T], error) {
	if config == nil {
		return nil, fmt.Errorf("local cache config is nil")
	}
	// 构建Ristretto缓存
	cache, err := ristretto.NewCache(&ristretto.Config[string, T]{
		NumCounters: config.Cache.NumCounters,
		MaxCost:     config.Cache.MaxCost,
		BufferItems: config.Cache.BufferItems,
	})
	if err != nil {
		return nil, fmt.Errorf("create ristretto cache failed: %w", err)
	}
	// 返回Ristretto缓存
	return &localCache[T]{local: cache, config: *config.Operation}, nil
}

func (l *localCache[T]) SetWithTTL(ctx context.Context, key string, value T, opts ...TTLOption) bool {
	ttl := l.ttlBuilder(opts...)
	isSuccess := l.local.SetWithTTL(key, value, 1, ttl.value)
	if !isSuccess {
		log.Printf("local set drop key: %s", key)
		return false
	}
	return true
}

func (l *localCache[T]) Get(ctx context.Context, key string) (T, bool) {
	value, isExist := l.local.Get(key)
	if !isExist {
		log.Printf("local get not exist key: %s", key)
		var zeroValue T
		return zeroValue, false
	}
	return value, true
}

func (l *localCache[T]) GetPointer(ctx context.Context, key string) (*T, bool) {
	value, isExist := l.local.Get(key)
	if !isExist {
		log.Printf("local get not exist key: %s", key)
		return nil, false
	}
	return &value, true
}

func (l *localCache[T]) Del(ctx context.Context, key string) {
	l.local.Del(key)
}

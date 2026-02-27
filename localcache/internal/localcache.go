package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/localcache/config"
	"github.com/LouYuanbo1/go-webservice/localcache/options"
	"github.com/dgraph-io/ristretto/v2"
)

type localCache[T any] struct {
	local  *ristretto.Cache[string, T]
	config *config.OperationConfig
}

func NewLocalCache[T any](config *config.LocalConfig) (*localCache[T], error) {
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
	return &localCache[T]{local: cache, config: config.Operation}, nil
}

func (l *localCache[T]) SetWithTTL(ctx context.Context, key string, value T, opts ...options.TTLOption) bool {
	ttl := options.TTLBuilder(l.config, opts...)
	isSuccess := l.local.SetWithTTL(key, value, 1, ttl.GetTTL())
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

package local

import (
	"context"
	"time"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/dgraph-io/ristretto/v2"
)

type localCache struct {
	local *ristretto.Cache[string, any]
}

func initLocalCache(config *Config) (*ristretto.Cache[string, any], error) {
	if config == nil {
		return nil, errorx.NewWithDetails(
			cache.ErrInit,
			"cache",
			"initLocalCache",
			"local cache config is nil",
			nil,
		)
	}
	// 构建Ristretto缓存
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: config.NumCounters,
		MaxCost:     config.MaxCost,
		BufferItems: config.BufferItems,
	})
	if err != nil {
		return nil, errorx.NewWithDetails(
			cache.ErrInit,
			"cache",
			"initLocalCache",
			"create ristretto cache failed",
			err,
		)
	}
	// 返回Ristretto缓存
	return ristrettoCache, nil
}

func newLocalCache(config *Config) (cache.Cache, error) {
	cache, err := initLocalCache(config)
	if err != nil {
		return nil, err
	}
	return &localCache{local: cache}, nil
}

func (lc *localCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	ok := lc.local.SetWithTTL(key, val, 1, ttl)
	if !ok {
		return errorx.New(
			cache.ErrSet,
			"cache",
			"Set",
			nil,
		)
	}
	return nil
}

func (lc *localCache) Get(ctx context.Context, key string, val any) error {
	val, ok := lc.local.Get(key)
	if !ok {
		return errorx.New(
			cache.ErrGet,
			"cache",
			"Get",
			nil,
		)
	}
	return nil
}

func (lc *localCache) Take(ctx context.Context, val any, key string, query func(val any) error, ttl time.Duration) error {
	err := lc.Get(ctx, key, val)
	if err != nil {
		if err := query(val); err != nil {
			return err
		}
		if err := lc.Set(ctx, key, val, ttl); err != nil {
			return err
		}
	}
	return nil
}

func (lc *localCache) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		lc.local.Del(key)
	}
	return nil
}

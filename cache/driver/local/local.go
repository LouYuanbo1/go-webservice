package local

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/LouYuanbo1/go-webservice/singleflightx"
	"github.com/coocood/freecache"
)

type localCache struct {
	local *freecache.Cache
	sf    singleflightx.SingleFlight
}

func initLocalCache(config *Config) (*freecache.Cache, error) {
	if config == nil {
		return nil, errorx.NewWithDetails(
			cache.ErrInit,
			"cache",
			"initLocalCache",
			"local cache config is nil",
			nil,
		)
	}
	localCache := freecache.NewCache(config.CacheSize)
	// 返回FreeCache缓存
	return localCache, nil
}

func newLocalCache(config *Config, sf singleflightx.SingleFlight) (cache.LocalCache, error) {
	cache, err := initLocalCache(config)
	if err != nil {
		return nil, err
	}
	return &localCache{local: cache, sf: sf}, nil
}

func (lc *localCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	byteKey := []byte(key)
	byteVal, err := json.Marshal(val)
	if err != nil {
		return errorx.New(
			cache.ErrJsonMarshal,
			"cache",
			"Set",
			nil,
		)
	}
	err = lc.local.Set(byteKey, byteVal, int(ttl))
	if err != nil {
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
	byteKey := []byte(key)
	jsonValue, err := lc.local.Get(byteKey)
	if err != nil {
		return errorx.New(
			cache.ErrGet,
			"cache",
			"Get",
			nil,
		)
	}
	err = json.Unmarshal(jsonValue, val)
	if err != nil {
		return errorx.New(
			cache.ErrJsonUnmarshal,
			"cache",
			"Get",
			nil,
		)
	}
	return nil
}

/*
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
*/

func (lc *localCache) Take(ctx context.Context, key string, val any, query func(val any) error, ttl time.Duration) error {
	// 使用 singleflight 保护整个回源过程
	data, fresh, err := lc.sf.DoEx(key, func() (any, error) {
		// 1. 尝试从缓存获取
		err := lc.Get(ctx, key, val)
		if err == nil {
			// 缓存命中，直接序列化 val 返回
			return json.Marshal(val)
		}

		// 2. 缓存未命中，执行回源查询,注意这里只有第一个 goroutine 会执行回源查询,其他 goroutine 会等待查询结果
		if err := query(val); err != nil {
			return nil, err
		}

		// 3. 回填缓存（异步或同步均可，此处同步）
		if err := lc.Set(ctx, key, val, ttl); err != nil {
			// 回填失败仅记录日志，不影响主流程
			// 可根据需要决定是否返回错误
			// 这里选择忽略，继续
			log.Printf("%v: %v", cache.WarnSetCacheFailed, err)
		}

		// 4. 返回序列化后的结果供其他等待者使用
		/*
			这里的序列化是防止多个 goroutine 同时获得同一个val指针,并发修改 val 导致数据不一致的问题
		*/
		return json.Marshal(val)
	})
	if err != nil {
		return err
	}

	// 如果当前 goroutine 是实际执行者（fresh == true），则 val 已经被填充且缓存已设，直接返回
	if fresh {
		return nil
	}

	// 等待者：将序列化数据反序列化到自己的 val 指针中,此时是val的副本
	return json.Unmarshal(data.([]byte), val)
}

func (lc *localCache) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		lc.local.Del([]byte(key))
	}
	return nil
}

func (lc *localCache) GetLocalCache() *freecache.Cache {
	return lc.local
}

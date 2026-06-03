package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/LouYuanbo1/go-webservice/singleflightx"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
	sf     singleflightx.SingleFlight
}

func InitRedisClient(config *Config, hooks ...redis.Hook) (*redis.Client, error) {
	if config == nil {
		return nil, errorx.NewWithDetails(
			cache.ErrInit,
			"cache",
			"initRedis",
			"RedisConfig cannot be nil",
			nil,
		)
	}
	// 构建Redis连接字符串
	redisAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	redisClient := redis.NewClient(&redis.Options{
		Addr:          redisAddr,
		Password:      config.Password,
		DB:            config.DB,
		Protocol:      config.Protocol,      // RESP3 协议,这个必须启用(2),否则在使用向量搜索时会出现无法寻找结果的问题
		UnstableResp3: config.UnstableResp3, // 启用 RESP3 支持
	})

	// 添加自定义钩子
	for _, hook := range hooks {
		redisClient.AddHook(hook)
	}

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, errorx.NewWithDetails(
			cache.ErrInit,
			"cache",
			"initRedis",
			"Redis connection failed",
			err,
		)
	}
	return redisClient, nil
}

func NewRedisCache(client *redis.Client, sf singleflightx.SingleFlight) (cache.RedisCache, error) {
	return &redisCache{client: client, sf: sf}, nil
}

func (rc *redisCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	jsonValue, err := json.Marshal(val)
	if err != nil {
		log.Printf("json marshal error: %v", err)
		return errorx.New(
			cache.ErrJsonMarshal,
			"cache",
			"Set",
			err,
		)
	}
	err = rc.client.Set(ctx, key, jsonValue, ttl).Err()
	if err != nil {
		log.Printf("redis set error: %v", err)
		return errorx.New(
			cache.ErrSet,
			"cache",
			"Set",
			err,
		)
	}
	return nil
}

/*
Get gets the cache with key and fills into val.
Val is the pointer to the value to fill.
*/
func (rc *redisCache) Get(ctx context.Context, key string, val any) error {
	jsonValue, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		return errorx.New(
			cache.ErrGet,
			"cache",
			"Get",
			err,
		)
	}
	err = json.Unmarshal(jsonValue, val)
	if err != nil {
		log.Printf("json unmarshal error: %v", err)
		return errorx.New(
			cache.ErrJsonUnmarshal,
			"cache",
			"Get",
			err,
		)
	}
	return nil
}

/*
Query is callback function to query DB.
If the cache is not found, it will call query to query from DB and set cache, then fill val.
Val is the pointer to the value to fill.
*/
/*
func (rc *redisCache) Take(ctx context.Context, val any, key string, query func(val any) error, ttl time.Duration) error {
	err := rc.Get(ctx, key, val)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return err
		}
		if err := query(val); err != nil {
			return err
		}
		if err := rc.Set(ctx, key, val, ttl); err != nil {
			return err
		}
	}
	return nil
}
*/

func (rc *redisCache) Take(ctx context.Context, key string, val any, query func(val any) error, ttl time.Duration) error {
	// 使用 singleflight 保护整个回源过程
	data, fresh, err := rc.sf.DoEx(key, func() (any, error) {
		// 1. 尝试从缓存获取
		err := rc.Get(ctx, key, val)
		if err == nil {
			// 缓存命中，直接序列化 val 返回
			return json.Marshal(val)
		}
		if !errors.Is(err, redis.Nil) {
			// 缓存错误（非不存在），直接返回错误
			return nil, err
		}

		// 2. 缓存未命中，执行回源查询,注意这里只有第一个 goroutine 会执行回源查询,其他 goroutine 会等待查询结果
		if err := query(val); err != nil {
			return nil, err
		}

		// 3. 回填缓存（异步或同步均可，此处同步）
		if err := rc.Set(ctx, key, val, ttl); err != nil {
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

func (rc *redisCache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	err := rc.client.Del(ctx, keys...).Err()
	if err != nil {
		log.Printf("redis del error: %v", err)
		return errorx.New(
			cache.ErrDel,
			"cache",
			"Del",
			err,
		)
	}
	return nil
}

func (rc *redisCache) GetRedisClient() *redis.Client {
	return rc.client
}

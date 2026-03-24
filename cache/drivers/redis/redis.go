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
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

func initRedis(config *Config) (*redis.Client, error) {
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

func newRedisCache(cfg *Config) (cache.Cache, error) {
	client, err := initRedis(cfg)
	if err != nil {
		return nil, err
	}
	return &redisCache{client: client}, nil
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

func (rc *redisCache) Del(ctx context.Context, keys ...string) error {
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

package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	// Set sets the cache with key and v, using c.expiry.
	Set(ctx context.Context, key string, val any) error
	SetWithExpire(ctx context.Context, key string, val any, opts ...TTLOption) error
	// Get gets the cache with key and fills into v.
	Get(ctx context.Context, key string, val any) error
	Take(ctx context.Context, val any, key string, query func(val any) error) error
	// TakeWithExpire takes the result from cache first, if not found,
	// query from DB and set cache using given expire, then return the result.
	TakeWithExpire(ctx context.Context, val any, key string,
		query func(val any) error, opts ...TTLOption) error
	// Del deletes cached values with keys.
	Del(ctx context.Context, keys ...string) error
}

type cache struct {
	client *redis.Client
	cfg    *OperationConfig
}

func NewCache(cfg *Config) Cache {
	client, err := InitRedis(cfg.DB)
	if err != nil {
		panic(err)
	}
	return &cache{client: client, cfg: cfg.Operation}
}

func (c *cache) Set(ctx context.Context, key string, val any) error {
	jsonValue, err := json.Marshal(val)
	if err != nil {
		return errorx.New(
			ErrJsonMarshal,
			"cache",
			"Set",
			err,
		)
	}
	err = c.client.Set(ctx, key, jsonValue, 0).Err()
	if err != nil {
		log.Printf("redis set error: %v", err)
		return errorx.New(
			ErrSet,
			"cache",
			"Set",
			err,
		)
	}
	return nil
}

func (c *cache) SetWithExpire(ctx context.Context, key string, val any, opts ...TTLOption) error {
	jsonValue, err := json.Marshal(val)
	if err != nil {
		log.Printf("json marshal error: %v", err)
		return errorx.New(
			ErrJsonMarshal,
			"cache",
			"Set",
			err,
		)
	}
	ttl := c.ttlBuilder(opts...)
	err = c.client.Set(ctx, key, jsonValue, ttl.value).Err()
	if err != nil {
		log.Printf("redis set error: %v", err)
		return errorx.New(
			ErrSet,
			"cache",
			"Set",
			err,
		)
	}
	return nil
}

func (c *cache) Get(ctx context.Context, key string, val any) error {
	jsonValue, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return errorx.New(
			ErrGet,
			"cache",
			"Get",
			err,
		)
	}
	err = json.Unmarshal(jsonValue, val)
	if err != nil {
		log.Printf("json unmarshal error: %v", err)
		return errorx.New(
			ErrJsonUnmarshal,
			"cache",
			"Get",
			err,
		)
	}
	return nil
}

// query is callback function to query DB
func (c *cache) Take(ctx context.Context, val any, key string, query func(val any) error) error {
	err := c.Get(ctx, key, val)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return err
		}
		if err := query(val); err != nil {
			return err
		}
		if err := c.Set(ctx, key, val); err != nil {
			return err
		}
	}
	return nil
}

func (c *cache) TakeWithExpire(ctx context.Context, val any, key string, query func(val any) error, opts ...TTLOption) error {
	err := c.Get(ctx, key, val)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return err
		}
		if err := query(val); err != nil {
			return err
		}
		if err := c.SetWithExpire(ctx, key, val, opts...); err != nil {
			return err
		}
	}
	return nil
}


func (c *cache) Del(ctx context.Context, keys ...string) error {
	err := c.client.Del(ctx, keys...).Err()
	if err != nil {
		log.Printf("redis del error: %v", err)
		return errorx.New(
			ErrDel,
			"cache",
			"Del",
			err,
		)
	}
	return nil
}
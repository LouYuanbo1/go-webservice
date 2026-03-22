package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	// Set sets the cache with key and v, using c.expiry.
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	// Get gets the cache with key and fills into v.
	Get(ctx context.Context, key string, val any) error
	// Take takes the result from cache first, if not found,
	// query from DB and set cache using given expire, then return the result.
	Take(ctx context.Context, key string, val any,
		query func(val any) error, ttl time.Duration) error
	// Del deletes cached values with keys.
	Del(ctx context.Context, keys ...string) error
}

type cache struct {
	client *redis.Client
}

func NewCache(cfg *Config) Cache {
	client, err := InitRedis(cfg.DB)
	if err != nil {
		panic(err)
	}
	return &cache{client: client}
}

func (c *cache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
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
	err = c.client.Set(ctx, key, jsonValue, ttl).Err()
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

/*
Get gets the cache with key and fills into val.
Val is the pointer to the value to fill.
*/
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

/*
Query is callback function to query DB.
If the cache is not found, it will call query to query from DB and set cache, then fill val.
Val is the pointer to the value to fill.
*/
func (c *cache) Take(ctx context.Context, key string, val any, query func(val any) error, ttl time.Duration) error {
	err := c.Get(ctx, key, val)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return err
		}
		if err := query(val); err != nil {
			return err
		}
		if err := c.Set(ctx, key, val, ttl); err != nil {
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

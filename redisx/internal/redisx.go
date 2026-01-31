package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisX[T any] struct {
	client        *redis.Client
	defaultTTLKey time.Duration
}

func NewRedisX[T any](client *redis.Client, defaultTTLKey time.Duration) *redisX[T] {
	return &redisX[T]{client: client, defaultTTLKey: defaultTTLKey}
}

func (rx *redisX[T]) SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		log.Printf("json marshal error: %v", err)
		return fmt.Errorf("json marshal error: %w", err)
	}
	err = rx.client.Set(ctx, key, jsonValue, ttl).Err()
	if err != nil {
		log.Printf("redis set error: %v", err)
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

func (rx *redisX[T]) SetWithDefaultTTL(ctx context.Context, key string, value T) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		log.Printf("json marshal error: %v", err)
		return fmt.Errorf("json marshal error: %w", err)
	}
	err = rx.client.Set(ctx, key, jsonValue, rx.defaultTTLKey).Err()
	if err != nil {
		log.Printf("redis set error: %v", err)
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

func (rx *redisX[T]) Get(ctx context.Context, key string) (T, error) {
	var result T
	jsonValue, err := rx.client.Get(ctx, key).Bytes()
	if err != nil {
		log.Printf("redis get error: %v", err)
		return result, fmt.Errorf("redis get error: %w", err)
	}
	err = json.Unmarshal(jsonValue, &result)
	if err != nil {
		log.Printf("json unmarshal error: %v", err)
		return result, fmt.Errorf("json unmarshal error: %w", err)
	}
	return result, nil
}

func (rx *redisX[T]) GetPointer(ctx context.Context, key string) (*T, error) {
	var result T
	jsonValue, err := rx.client.Get(ctx, key).Bytes()
	if err != nil {
		log.Printf("redis get error: %v", err)
		return nil, fmt.Errorf("redis get error: %w", err)
	}
	err = json.Unmarshal(jsonValue, &result)
	if err != nil {
		log.Printf("json unmarshal error: %v", err)
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}
	return &result, nil
}

func (rx *redisX[T]) Del(ctx context.Context, key string) error {
	err := rx.client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("redis del error: %v", err)
		return fmt.Errorf("redis del error: %w", err)
	}
	return nil
}

func (rx *redisX[T]) Acquire(ctx context.Context, key string, expiration time.Duration) (string, bool, error) {
	lockID := uuid.New().String()
	success, err := rx.client.SetNX(ctx, key, lockID, expiration).Result()
	if err != nil {
		return "", false, err
	}
	return lockID, success, nil
}

func (rx *redisX[T]) Release(ctx context.Context, key string, lockID string) error {
	luaScript := `
    if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("del", KEYS[1])
    else
        return 0
    end
    `
	script := redis.NewScript(luaScript)
	_, err := script.Run(ctx, rx.client, []string{key}, lockID).Result()
	if err != nil {
		log.Printf("redis unlock error: %v", err)
		return fmt.Errorf("redis unlock error: %w", err)
	}
	return nil
}

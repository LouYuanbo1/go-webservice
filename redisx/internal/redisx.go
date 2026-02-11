package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/LouYuanbo1/go-webservice/redisx/options"
	"github.com/go-viper/mapstructure/v2"
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

func (rx *redisX[T]) SetWithTTL(ctx context.Context, key string, value T, opts ...options.TTLOption) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		log.Printf("json marshal error: %v", err)
		return fmt.Errorf("json marshal error: %w", err)
	}

	ttl := rx.ttlBuilder(opts...)

	err = rx.client.Set(ctx, key, jsonValue, ttl.GetTTL()).Err()
	if err != nil {
		log.Printf("redis set error: %v", err)
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

/*
对于结构体使用HSetWithTTL方法时,需要注意结构体字段的tag中是否包含了redis标签,
或实现encoding.BinaryMarshaler 接口.
详情:https://pkg.go.dev/github.com/redis/go-redis/v9@v9.17.3#Client.HSet

For using HSetWithTTL method with struct, you need to ensure that the struct fields have redis tag,
or implement encoding.BinaryMarshaler interface.
Details:https://pkg.go.dev/github.com/redis/go-redis/v9@v9.17.3#Client.HSet

For example:

	type User struct {
		ID        uint64    `gorm:"primaryKey" redis:"id"`
		Name      string    `gorm:"not null" redis:"name"`
		Email     string    `gorm:"not null;unique" redis:"email"`
		CreatedAt time.Time `gorm:"not null;default:current_timestamp"`
		UpdatedAt time.Time `gorm:"not null;default:current_timestamp"`
	}
*/
func (rx *redisX[T]) HSetWithTTL(ctx context.Context, key string, value T, opts ...options.TTLOption) error {
	err := rx.client.HSet(ctx, key, value).Err()
	if err != nil {
		log.Printf("redis hset error: %v", err)
		return fmt.Errorf("redis hset error: %w", err)
	}

	ttl := rx.ttlBuilder(opts...)

	err = rx.client.Expire(ctx, key, ttl.GetTTL()).Err()
	if err != nil {
		log.Printf("redis expire error: %v", err)
		return fmt.Errorf("redis expire error: %w", err)
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

func (rx *redisX[T]) HGet(ctx context.Context, key string, field string) (string, error) {
	result, err := rx.client.HGet(ctx, key, field).Result()
	if err != nil {
		log.Printf("redis hget error: %v", err)
		return result, fmt.Errorf("redis hget error: %w", err)
	}
	return result, nil
}

func (rx *redisX[T]) HMGet(ctx context.Context, key string, fields ...string) ([]any, error) {
	result, err := rx.client.HMGet(ctx, key, fields...).Result()
	if err != nil {
		log.Printf("redis hmget error: %v", err)
		return nil, fmt.Errorf("redis hmget error: %w", err)
	}
	return result, nil
}

func (rx *redisX[T]) HGetAll(ctx context.Context, key string) (T, error) {
	var result T
	resultMap, err := rx.client.HGetAll(ctx, key).Result()
	if err != nil {
		log.Printf("redis hget error: %v", err)
		return result, fmt.Errorf("redis hget error: %w", err)
	}
	config := &mapstructure.DecoderConfig{
		TagName:          "redis", // 匹配结构体的redis标签
		Result:           &result, // 绑定的目标结构体（必须传指针）
		WeaklyTypedInput: true,    // 开启弱类型自动转换（string→int/bool/float等）
		ZeroFields:       true,    // 绑定前先将结构体置为零值（可选，默认true）
	}
	// 创建解码器并执行转换
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		panic(fmt.Sprintf("创建解码器失败：%v", err))
	}
	err = decoder.Decode(resultMap)
	if err != nil {
		log.Printf("mapstructure decode error: %v", err)
		return result, fmt.Errorf("mapstructure decode error: %w", err)
	}
	return result, nil
}

func (rx *redisX[T]) HGetAllPointer(ctx context.Context, key string) (*T, error) {
	var result T
	resultMap, err := rx.client.HGetAll(ctx, key).Result()
	if err != nil {
		log.Printf("redis hget error: %v", err)
		return nil, fmt.Errorf("redis hget error: %w", err)
	}
	config := &mapstructure.DecoderConfig{
		TagName:          "redis", // 匹配结构体的redis标签
		Result:           &result, // 绑定的目标结构体（必须传指针）
		WeaklyTypedInput: true,    // 开启弱类型自动转换（string→int/bool/float等）
		ZeroFields:       true,    // 绑定前先将结构体置为零值（可选，默认true）
	}
	// 创建解码器并执行转换
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		panic(fmt.Sprintf("创建解码器失败：%v", err))
	}
	err = decoder.Decode(resultMap)
	if err != nil {
		log.Printf("mapstructure decode error: %v", err)
		return nil, fmt.Errorf("mapstructure decode error: %w", err)
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

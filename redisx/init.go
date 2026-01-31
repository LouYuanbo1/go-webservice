package redisx

import (
	"context"
	"fmt"

	"github.com/LouYuanbo1/go-webservice/redisx/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis(config *config.RedisConfig) (*redis.Client, error) {
	if config == nil {
		return nil, fmt.Errorf("RedisConfig cannot be nil")
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
		return nil, fmt.Errorf("Redis connection failed: %w", err)
	}
	return redisClient, nil
}

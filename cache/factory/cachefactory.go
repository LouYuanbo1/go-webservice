package cachefactory

import (
	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/cache/internal/local"
	"github.com/LouYuanbo1/go-webservice/cache/internal/redis"
	"github.com/LouYuanbo1/go-webservice/errorx"
)

func NewCache(config *cache.Config) (cache.Cache, error) {
	switch config.Type {
	case "local":
		return local.NewLocalCache(config.Local)
	case "redis":
		return redis.NewRedisCache(config.Redis)
	default:
		return nil, errorx.NewWithDetails(
			cache.ErrInit,
			"cache",
			"NewCache",
			"cache cache type is not supported",
			nil,
		)
	}
}

package localcache

import (
	"context"
	"time"

	"github.com/LouYuanbo1/go-webservice/localcache/config"
	"github.com/LouYuanbo1/go-webservice/localcache/internal"
)

type LocalCache[T any] interface {
	// Set sets the value for the given key.
	SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) bool
	// SetWithDefaultTTL sets the value for the given key with the default expiration time.
	SetWithDefaultTTL(ctx context.Context, key string, value T) bool

	// Get gets the value for the given key.
	Get(ctx context.Context, key string) (T, bool)
	// GetPointer gets the pointer value for the given key.
	GetPointer(ctx context.Context, key string) (*T, bool)

	// Delete deletes the value for the given key.
	Del(ctx context.Context, key string)
}

func NewLocalCache[T any](config *config.LocalConfig) (LocalCache[T], error) {
	return internal.NewLocalCache[T](config)
}

package cachedb

import "time"

type ttl struct {
	value time.Duration
}

type TTLOption func(*ttl)

func WithTTL(value time.Duration) TTLOption {
	return func(t *ttl) {
		t.value = value
	}
}

func BuildTTL(cfg *Config, opts ...TTLOption) time.Duration {
	tt := &ttl{}
	if cfg != nil {
		tt.value = cfg.TTL
	}
	for _, opt := range opts {
		opt(tt)
	}
	return tt.value
}

// SafeGap 返回索引缓存与主键缓存的安全间隔；cfg 为空时取 0，避免空指针。
func SafeGap(cfg *Config) time.Duration {
	if cfg == nil {
		return 0
	}
	return cfg.CacheSafeGapBetweenIndexAndPrimary
}

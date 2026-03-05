package redisx

import (
	"time"
)

type ttl struct {
	value time.Duration
}

func newTTLByConfig(cfg OperationConfig) *ttl {
	return &ttl{value: time.Duration(cfg.TTL) * time.Second}
}

type TTLOption func(*ttl)

func WithTTL(value time.Duration) TTLOption {
	return func(t *ttl) {
		t.value = value
	}
}

func (r *redisX[T]) ttlBuilder(opts ...TTLOption) *ttl {
	t := newTTLByConfig(r.config)
	for _, opt := range opts {
		opt(t)
	}
	return t
}

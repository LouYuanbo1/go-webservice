package options

import (
	"time"

	"github.com/LouYuanbo1/go-webservice/localcache/config"
)

type ttl struct {
	value time.Duration
}

func (t *ttl) GetTTL() time.Duration {
	return t.value
}

func NewTTL() *ttl {
	return &ttl{}
}

func NewTTLByConfig(cfg *config.OperationConfig) *ttl {
	return &ttl{value: time.Duration(cfg.TTL) * time.Second}
}

func TTLBuilder(cfg *config.OperationConfig, opts ...TTLOption) *ttl {
	t := NewTTLByConfig(cfg)
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t *ttl) WithTTL(ttl time.Duration) *ttl {
	t.value = ttl
	return t
}

type TTLOption func(*ttl)

func WithTTL(value time.Duration) TTLOption {
	return func(t *ttl) {
		t.value = value
	}
}

func NewTTLWithOptions(opts ...TTLOption) *ttl {
	t := NewTTL()
	for _, opt := range opts {
		opt(t)
	}
	return t
}

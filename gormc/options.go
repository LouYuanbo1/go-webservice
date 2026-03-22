package gormc

import (
	"time"
)

type TTLProvider interface {
	GetTTL() time.Duration
}

type ttl struct {
	value time.Duration
}

func (t *ttl) GetTTL() time.Duration {
	return t.value
}

type TTLOption func(*ttl)

func WithTTL(value time.Duration) TTLOption {
	return func(t *ttl) {
		t.value = value
	}
}

func TTLBuilder(defaultTTL time.Duration, opts ...TTLOption) TTLProvider {
	t := &ttl{value: defaultTTL}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

package options

import "time"

type TTL struct {
	ttl time.Duration
}

func NewTTL() *TTL {
	return &TTL{}
}

func (t *TTL) GetTTL() time.Duration {
	return t.ttl
}

func (t *TTL) WithTTL(ttl time.Duration) *TTL {
	t.ttl = ttl
	return t
}

type TTLOption func(*TTL)

func WithTTL(ttl time.Duration) TTLOption {
	return func(t *TTL) {
		t.ttl = ttl
	}
}

func NewTTLWithOptions(opts ...TTLOption) *TTL {
	t := NewTTL()
	for _, opt := range opts {
		opt(t)
	}
	return t
}

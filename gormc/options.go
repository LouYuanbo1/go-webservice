package gormc

import (
	"time"
)

type ttl struct {
	value time.Duration
}

type TTLOption func(*ttl)

func WithTTL(value time.Duration) TTLOption {
	return func(t *ttl) {
		t.value = value
	}
}

func (cdb *cacheDB) ttlBuilder(opts ...TTLOption) *ttl {
	t := &ttl{value: cdb.cfg.TTL}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (tcdb *typedCacheDB[T, ID, PT]) ttlBuilder(opts ...TTLOption) *ttl {
	t := &ttl{value: tcdb.cfg.TTL}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

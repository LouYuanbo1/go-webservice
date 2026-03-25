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

func (cc *cachedConn) ttlBuilder(opts ...TTLOption) *ttl {
	t := &ttl{value: cc.cfg.TTL}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (tcc *typedCachedConn[T, ID, PT]) ttlBuilder(opts ...TTLOption) *ttl {
	t := &ttl{value: tcc.cfg.TTL}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

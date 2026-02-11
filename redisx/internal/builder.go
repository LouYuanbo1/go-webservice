package internal

import (
	"github.com/LouYuanbo1/go-webservice/redisx/options"
)

func (rx *redisX[T]) ttlBuilder(opts ...options.TTLOption) *options.TTL {
	ttl := options.NewTTL().WithTTL(rx.defaultTTLKey)
	for _, opt := range opts {
		opt(ttl)
	}
	return ttl
}

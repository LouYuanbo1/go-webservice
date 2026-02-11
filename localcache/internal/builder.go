package internal

import (
	"github.com/LouYuanbo1/go-webservice/localcache/options"
)

func (l *localCache[T]) ttlBuilder(opts ...options.TTLOption) *options.TTL {
	ttl := options.NewTTL().WithTTL(l.defaultTTLKey)
	for _, opt := range opts {
		opt(ttl)
	}
	return ttl
}

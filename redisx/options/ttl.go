package options

import "time"

type TTL struct {
	Value time.Duration
}

type TTLOption func(*TTL)

func WithTTL(ttl time.Duration) TTLOption {
	return func(t *TTL) {
		t.Value = ttl
	}
}

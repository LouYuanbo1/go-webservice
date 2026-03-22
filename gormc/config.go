package gormc

import "time"

type Config struct {
	TTL                                time.Duration `mapstructure:"ttl"`
	CacheSafeGapBetweenIndexAndPrimary time.Duration `mapstructure:"cache_safe_gap_between_index_and_primary"`
}

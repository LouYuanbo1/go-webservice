package localcache

type Config struct {
	Cache     *CacheConfig     `mapstructure:"cache"`
	Operation *OperationConfig `mapstructure:"operation"`
}

type CacheConfig struct {
	Type        string `mapstructure:"type"`
	NumCounters int64  `mapstructure:"num_counters"`
	MaxCost     int64  `mapstructure:"max_cost"`
	BufferItems int64  `mapstructure:"buffer_items"`
}

type OperationConfig struct {
	TTL int64 `mapstructure:"ttl"`
}

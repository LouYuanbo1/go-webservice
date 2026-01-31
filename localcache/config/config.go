package config

type LocalConfig struct {
	Type        string `mapstructure:"type"`
	NumCounters int64  `mapstructure:"num_counters"`
	MaxCost     int64  `mapstructure:"max_cost"`
	BufferItems int64  `mapstructure:"buffer_items"`
	DefaultTTL  int64  `mapstructure:"default_ttl"`
}

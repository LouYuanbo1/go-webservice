package cache

type Config struct {
	Type string `mapstructure:"type"`
	Redis *RedisConfig `mapstructure:"redis"`
	Local *LocalConfig `mapstructure:"local"`
}

// RedisConfig is the configuration for redis
type RedisConfig struct {
	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	Password      string `mapstructure:"password"`
	DB            int    `mapstructure:"db"`
	Protocol      int    `mapstructure:"protocol"`
	UnstableResp3 bool   `mapstructure:"unstable_resp3"`
}

type LocalConfig struct {
	NumCounters int64 `mapstructure:"num_counters"`
	MaxCost     int64 `mapstructure:"max_cost"`
	BufferItems int64 `mapstructure:"buffer_items"`
}

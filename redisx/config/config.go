package config

type RedisConfig struct {
	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	Password      string `mapstructure:"password"`
	DB            int    `mapstructure:"db"`
	Protocol      int    `mapstructure:"protocol"`
	UnstableResp3 bool   `mapstructure:"unstable_resp3"`
	MaxSize       int    `mapstructure:"max_size"`
	DefaultTTL    int64  `mapstructure:"default_ttl"`
}

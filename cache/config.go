package cache

type Config struct {
	DB *DBConfig `mapstructure:"db"`
}

// RedisDBConfig is the configuration for redis db
type DBConfig struct {
	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	Password      string `mapstructure:"password"`
	DB            int    `mapstructure:"db"`
	Protocol      int    `mapstructure:"protocol"`
	UnstableResp3 bool   `mapstructure:"unstable_resp3"`
}

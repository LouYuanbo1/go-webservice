package redisx

type Config struct {
	DB        *DBConfig        `mapstructure:"db"`
	Operation *OperationConfig `mapstructure:"operation"`
}

type DBConfig struct {
	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	Password      string `mapstructure:"password"`
	DB            int    `mapstructure:"db"`
	Protocol      int    `mapstructure:"protocol"`
	UnstableResp3 bool   `mapstructure:"unstable_resp3"`
}

type OperationConfig struct {
	TTL int64 `mapstructure:"ttl"`
}

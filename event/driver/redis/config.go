package redis

// Config 是 Redis 事件总线驱动的连接配置。
type Config struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password,omitempty"`
	DB       int    `mapstructure:"db,omitempty"`
}

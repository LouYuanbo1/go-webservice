package config

type DBConfig struct {
	Type     string `mapstructure:"type"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	// 最大打开连接数 (建议值: 25)
	MaxOpenConns int `mapstructure:"max_open_conns"`
	// 最大空闲连接数 (建议值: 25)
	MaxIdleConns int `mapstructure:"max_idle_conns"`
	// 连接最大生命周期 (建议值: 5m)
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
	// 时区配置 (e.g. "Asia/Shanghai")
	TimeZone string `mapstructure:"time_zone"`
	// 日志级别 Silent=1, Warn=2, Error=3, Info=4 (默认)
	LogLevel int `mapstructure:"log_level"`
	// schema.sql 文件路径
	SchemaFile string `mapstructure:"schema_file"`

	MySQL    MySQL    `mapstructure:"mysql"`
	Postgres Postgres `mapstructure:"postgres"`
}

type MySQL struct {
	TLS string `mapstructure:"tls"`
}

type Postgres struct {
	SSLMode string `mapstructure:"ssl_mode"`
}

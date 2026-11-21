package rabbitmq

// Config 是 RabbitMQ 事件总线驱动的连接配置。
type Config struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	VHost    string `mapstructure:"vhost,omitempty"`
	// Exchange 是事件交换机名称，topic 类型，默认 event.bus。
	Exchange string `mapstructure:"exchange"`
}

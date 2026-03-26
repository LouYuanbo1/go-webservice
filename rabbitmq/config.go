package rabbitmq

import (
	"fmt"
)

type RabbitConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	VHost    string `mapstructure:"vhost,omitempty"`
}

type RabbitConsumerConfig struct {
	RabbitConfig
	ListenerQueues []ConsumerConfig `mapstructure:"listener_queues"`
}

type ConsumerConfig struct {
	Name      string `mapstructure:"name"`
	AutoAck   bool   `mapstructure:"auto_ack"`
	Exclusive bool `mapstructure:"exclusive,omitempty"`
	// Set to true, which means that messages sent by producers in the same connection
	// cannot be delivered to consumers in this connection.
	NoLocal bool `mapstructure:"no_local,omitempty"`
	// Whether to block processing
	NoWait bool `mapstructure:"no_wait,omitempty"`
}

type RabbitProducerConfig struct {
	RabbitConfig
	ContentType string `mapstructure:"content_type"`
}

type QueueConfig struct {
	Name       string `mapstructure:"name"`
	Durable    bool `mapstructure:"durable,omitempty"`
	AutoDelete bool `mapstructure:"auto_delete,omitempty"`
	Exclusive  bool `mapstructure:"exclusive,omitempty"`
	NoWait     bool `mapstructure:"no_wait,omitempty"`
}

type ExchangeConfig struct {
	ExchangeName string
	Kind         string        `mapstructure:"kind"` // exchange type
	Durable      bool          `mapstructure:"durable,omitempty"`
	AutoDelete   bool          `mapstructure:"auto_delete,omitempty"`
	Internal     bool          `mapstructure:"internal,omitempty"`
	NoWait       bool          `mapstructure:"no_wait,omitempty"`
	Queues       []QueueConfig `mapstructure:"queues"`
}

type BindConfig struct {
	QueueName string `mapstructure:"queue_name"`
	RouteKey  string `mapstructure:"route_key"`
	Exchange  string `mapstructure:"exchange"`
	NotWait   bool   `mapstructure:"not_wait,omitempty"`
}

func getRabbitURL(cfg RabbitConfig) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", cfg.Username, cfg.Password,
		cfg.Host, cfg.Port, cfg.VHost)
}

package rabbitmq

import (
	"fmt"
)

type RabbitConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	VHost    string `json:"vhost,omitempty"`
}

type RabbitConsumerConfig struct {
	RabbitConfig
	ListenerQueues []ConsumerConfig
}

type ConsumerConfig struct {
	Name      string
	AutoAck   bool `json:"auto_ack,omitempty"`
	Exclusive bool `json:"exclusive,omitempty"`
	// Set to true, which means that messages sent by producers in the same connection
	// cannot be delivered to consumers in this connection.
	NoLocal bool `json:"no_local,omitempty"`
	// Whether to block processing
	NoWait bool `json:"no_wait,omitempty"`
}

type RabbitProducerConfig struct {
	RabbitConfig
	ContentType string `json:"content_type,omitempty"` // MIME content type
}

type QueueConfig struct {
	Name       string
	Durable    bool `json:"durable,omitempty"`
	AutoDelete bool `json:"auto_delete,omitempty"`
	Exclusive  bool `json:"exclusive,omitempty"`
	NoWait     bool `json:"no_wait,omitempty"`
}

type ExchangeConfig struct {
	ExchangeName string
	Kind         string        `json:"kind,omitempty"` // exchange type
	Durable      bool          `json:"durable,omitempty"`
	AutoDelete   bool          `json:"auto_delete,omitempty"`
	Internal     bool          `json:"internal,omitempty"`
	NoWait       bool          `json:"no_wait,omitempty"`
	Queues       []QueueConfig `json:"queues,omitempty"`
}

type BindConfig struct {
	QueueName string
	RouteKey  string
	Exchange  string
	NotWait   bool
}

func getRabbitURL(cfg RabbitConfig) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", cfg.Username, cfg.Password,
		cfg.Host, cfg.Port, cfg.VHost)
}

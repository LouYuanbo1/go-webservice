package rabbitmq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRabbitURL(t *testing.T) {
	tests := []struct {
		name     string
		cfg      RabbitConfig
		expected string
	}{
		{
			name: "default vhost",
			cfg: RabbitConfig{
				Username: "guest",
				Password: "guest",
				Host:     "localhost",
				Port:     5672,
				VHost:    "",
			},
			expected: "amqp://guest:guest@localhost:5672/",
		},
		{
			name: "custom vhost",
			cfg: RabbitConfig{
				Username: "user",
				Password: "pass",
				Host:     "rabbit.example.com",
				Port:     5671,
				VHost:    "my_vhost",
			},
			expected: "amqp://user:pass@rabbit.example.com:5671/my_vhost",
		},
		{
			name: "empty vhost",
			cfg: RabbitConfig{
				Username: "admin",
				Password: "secret",
				Host:     "127.0.0.1",
				Port:     5672,
				VHost:    "",
			},
			expected: "amqp://admin:secret@127.0.0.1:5672/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRabbitURL(tt.cfg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewProducer_InvalidConfig(t *testing.T) {
	cfg := RabbitProducerConfig{
		RabbitConfig: RabbitConfig{
			Username: "guest",
			Password: "guest",
			Host:     "invalid-host",
			Port:     5672,
			VHost:    "/",
		},
		ContentType: "application/json",
	}

	producer, err := NewProducer(cfg)

	assert.Nil(t, producer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrConnectFailed.Error())
}

func TestNewConsumer_InvalidConfig(t *testing.T) {
	cfg := RabbitConsumerConfig{
		RabbitConfig: RabbitConfig{
			Username: "guest",
			Password: "guest",
			Host:     "invalid-host",
			Port:     5672,
			VHost:    "/",
		},
		ListenerQueues: []ConsumerConfig{
			{
				Name:    "test-queue",
				AutoAck: true,
			},
		},
	}

	handler := func(msg []byte) error {
		return nil
	}

	consumer, err := NewConsumer(cfg, handler)

	assert.Nil(t, consumer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrConnectFailed.Error())
}

func TestProducerInterface(t *testing.T) {
	var p Producer = &producer{}
	assert.NotNil(t, p)
}

func TestConsumerInterface(t *testing.T) {
	handler := func(msg []byte) error { return nil }
	var c Consumer = &consumer{handler: handler}
	assert.NotNil(t, c)
}

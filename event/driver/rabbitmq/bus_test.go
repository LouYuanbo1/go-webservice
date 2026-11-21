package rabbitmq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRabbitURL(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		expected string
	}{
		{
			name:     "default vhost",
			cfg:      Config{Username: "guest", Password: "guest", Host: "localhost", Port: 5672},
			expected: "amqp://guest:guest@localhost:5672/",
		},
		{
			name:     "slash vhost",
			cfg:      Config{Username: "guest", Password: "guest", Host: "localhost", Port: 5672, VHost: "/"},
			expected: "amqp://guest:guest@localhost:5672/",
		},
		{
			name:     "custom vhost",
			cfg:      Config{Username: "user", Password: "pass", Host: "rabbit.example.com", Port: 5671, VHost: "my_vhost"},
			expected: "amqp://user:pass@rabbit.example.com:5671/my_vhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, rabbitURL(tt.cfg))
		})
	}
}

func TestNewBus_NilConfig(t *testing.T) {
	bus, err := NewBus(nil)
	assert.Nil(t, bus)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrConnectFailed.Error())
}

func TestNewBus_InvalidHost(t *testing.T) {
	bus, err := NewBus(&Config{
		Username: "guest",
		Password: "guest",
		Host:     "127.0.0.1",
		Port:     1, // 本地封闭端口，连接立即失败
	})
	assert.Nil(t, bus)
	assert.Error(t, err)
}

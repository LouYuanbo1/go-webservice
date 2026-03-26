package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type Admin interface {
	DeclareExchange(cfg ExchangeConfig, args amqp091.Table) error
	DeclareQueue(cfg QueueConfig, args amqp091.Table) error
	Bind(cfg BindConfig, args amqp091.Table) error
}

type admin struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

func NewAdmin(cfg RabbitConfig) (Admin, error) {
	var admin admin
	conn, err := amqp091.Dial(getRabbitURL(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to connect rabbitmq, error: %v", err)
	}

	admin.conn = conn
	channel, err := admin.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel, error: %v", err)
	}

	admin.channel = channel
	return &admin, nil
}

/*
生产者 --(路由键)--> 交换机 --(匹配绑定键)--> 队列
                    ↑
                    (交换机定义)
*/

func (a *admin) DeclareExchange(cfg ExchangeConfig, args amqp091.Table) error {
	return a.channel.ExchangeDeclare(
		cfg.ExchangeName,
		cfg.Kind,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Internal,
		cfg.NoWait,
		args,
	)
}

/*
生产者 --(路由键)--> 交换机 --(匹配绑定键)--> 队列
                    						↑
                    						(队列定义)
*/

func (a *admin) DeclareQueue(cfg QueueConfig, args amqp091.Table) error {
	_, err := a.channel.QueueDeclare(
		cfg.Name,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Exclusive,
		cfg.NoWait,
		args,
	)

	return err
}

/*
生产者 --(路由键)--> 交换机 --(匹配绑定键)--> 队列
                          ↑
                       绑定键(绑定定义)
*/

func (a *admin) Bind(cfg BindConfig, args amqp091.Table) error {
	return a.channel.QueueBind(
		cfg.QueueName,
		cfg.RouteKey,
		cfg.Exchange,
		cfg.NotWait,
		args,
	)
}

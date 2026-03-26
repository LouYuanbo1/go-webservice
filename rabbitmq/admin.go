package rabbitmq

import (
	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/rabbitmq/amqp091-go"
)

type Admin interface {
	ExchangeDeclare(cfg ExchangeConfig, args amqp091.Table) error
	QueueDeclare(cfg QueueConfig, args amqp091.Table) error
	QueueBind(cfg BindConfig, args amqp091.Table) error
}

type admin struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

func NewAdmin(cfg RabbitConfig) (Admin, error) {
	var admin admin
	conn, err := amqp091.Dial(getRabbitURL(cfg))
	if err != nil {
		return nil, errorx.New(
			ErrConnectFailed,
			"rabbitmq",
			"NewAdmin",
			err,
		)
	}
	admin.conn = conn
	channel, err := admin.conn.Channel()
	if err != nil {
		return nil, errorx.New(
			ErrChannelFailed,
			"rabbitmq",
			"NewAdmin",
			err,
		)
	}
	admin.channel = channel
	return &admin, nil
}

/*
生产者 --(路由键)--> 交换机 --(匹配绑定键)--> 队列
                    ↑
                    (交换机定义)
*/

func (a *admin) ExchangeDeclare(cfg ExchangeConfig, args amqp091.Table) error {
	err:= a.channel.ExchangeDeclare(
		cfg.ExchangeName,
		cfg.Kind,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Internal,
		cfg.NoWait,
		args,
	)
	if err != nil {
		return errorx.New(
			ErrExchangeDeclareFailed,
			"rabbitmq",
			"ExchangeDeclare",
			err,
		)
	}
	return nil
}

/*
生产者 --(路由键)--> 交换机 --(匹配绑定键)--> 队列
                    						↑
                    						(队列定义)
*/

func (a *admin) QueueDeclare(cfg QueueConfig, args amqp091.Table) error {
	_, err := a.channel.QueueDeclare(
		cfg.Name,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Exclusive,
		cfg.NoWait,
		args,
	)
	if err != nil {
		return errorx.New(
			ErrQueueDeclareFailed,
			"rabbitmq",
			"QueueDeclare",
			err,
		)
	}
	return nil
}

/*
生产者 --(路由键)--> 交换机 --(匹配绑定键)--> 队列
                          ↑
                       绑定键(绑定定义)
*/

func (a *admin) QueueBind(cfg BindConfig, args amqp091.Table) error {
	err := a.channel.QueueBind(
		cfg.QueueName,
		cfg.RouteKey,
		cfg.Exchange,
		cfg.NotWait,
		args,
	)
	if err != nil {
		return errorx.New(
			ErrQueueBindFailed,
			"rabbitmq",
			"QueueBind",
			err,
		)
	}
	return nil
}

package rabbitmq

import (
	"log"
	"sync"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/rabbitmq/amqp091-go"
)

type ConsumeHandler interface {
	Consume(message string) error
}

type Consumer interface {
	Start()
	Stop()
}

type consumer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	wg      sync.WaitGroup
	handler ConsumeHandler
	queues  []ConsumerConfig
}

func NewConsumer(consumerConfig RabbitConsumerConfig, handler ConsumeHandler) (Consumer, error) {
	c := &consumer{handler: handler, queues: consumerConfig.ListenerQueues}
	conn, err := amqp091.Dial(getRabbitURL(consumerConfig.RabbitConfig))
	if err != nil {
		return nil, errorx.New(
			ErrConnectFailed,
			"rabbitmq",
			"NewConsumer",
			err,
		)
	}

	c.conn = conn
	channel, err := c.conn.Channel()
	if err != nil {
		return nil, errorx.New(
			ErrChannelFailed,
			"rabbitmq",
			"NewConsumer",
			err,
		)
	}

	c.channel = channel
	return c, nil
}

func (c *consumer) Start() {
	for _, que := range c.queues {
		msg, err := c.channel.Consume(
			que.Name,
			"",
			que.AutoAck,
			que.Exclusive,
			que.NoLocal,
			que.NoWait,
			nil,
		)
		if err != nil {
			e := errorx.New(
				ErrConsumeFailed,
				"rabbitmq",
				"Start",
				err,
			)
			log.Fatalf("%v", e)
		}
		c.wg.Go(func() {
			for d := range msg {
				if err := c.handler.Consume(string(d.Body)); err != nil {
					e := errorx.New(
						ErrConsumeFailed,
						"rabbitmq",
						"Start",
						err,
					)
					log.Printf("Error consuming: %s, %v", string(d.Body), e)
				}
			}
		})
	}
	c.wg.Wait()
}

func (c *consumer) Stop() {
	c.channel.Close()
	c.conn.Close()
}

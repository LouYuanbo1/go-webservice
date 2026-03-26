package rabbitmq

import (
	"fmt"
	"log"
	"sync"

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
		return nil, fmt.Errorf("failed to connect rabbitmq, error: %v", err)
	}

	c.conn = conn
	channel, err := c.conn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel: %v", err)
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
			log.Fatalf("failed to listener, error: %v", err)
		}
		c.wg.Go(func() {
			for d := range msg {
				if err := c.handler.Consume(string(d.Body)); err != nil {
					log.Printf("Error on consuming: %s, error: %v", string(d.Body), err)
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

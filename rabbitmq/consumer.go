package rabbitmq

import (
	"log"
	"sync"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/rabbitmq/amqp091-go"
)

type Consumer interface {
	Start()
	Stop()
}

type consumer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	wg      sync.WaitGroup
	handler func(msg string) error
	queues  []ConsumerConfig
}

func NewConsumer(consumerConfig RabbitConsumerConfig, handler func(msg string) error) (Consumer, error) {
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

/*
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
				if err := c.handler(string(d.Body)); err != nil {
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
*/

func (c *consumer) Start() {
	for _, que := range c.queues {
		// 启动消费者（autoAck 仍需要作为参数传入，但逻辑分支已在外部确定）
		msgs, err := c.channel.Consume(
			que.Name,
			"",          // consumer tag
			que.AutoAck, // 仅用于告诉 RabbitMQ 是否自动确认，消费逻辑在分支中处理
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

		c.wg.Add(1)

		// 根据 autoAck 值在外部决定启动哪种 goroutine，内部不再访问 que 变量
		if que.AutoAck {
			// 自动确认模式：goroutine 仅负责处理业务，无需手动 Ack
			go func(queueName string, deliveries <-chan amqp091.Delivery) {
				defer c.wg.Done()
				for d := range deliveries {
					if err := c.handler(string(d.Body)); err != nil {
						log.Printf("[queue:%s] Error consuming (auto ack): %s, %v", queueName, string(d.Body), err)
					}
				}
			}(que.Name, msgs)
		} else {
			// 手动确认模式：goroutine 需要显式 Ack / Nack
			go func(queueName string, deliveries <-chan amqp091.Delivery) {
				defer c.wg.Done()
				for d := range deliveries {
					// 使用匿名函数确保 panic 恢复不影响循环
					func() {
						defer func() {
							if r := recover(); r != nil {
								log.Printf("[queue:%s] Panic consuming: %v", queueName, r)
								/*
									multiple：同 Ack，表示是否批量拒绝。
									false：仅拒绝当前消息。
									true：拒绝当前及之前所有未被确认的消息。
									requeue：是否将消息重新放回队列。
									true：消息被重新排队，可能被再次投递给消费者（常用于临时失败重试）。
									false：消息被直接丢弃，或根据队列配置进入死信交换器（DLX）。
								*/
								if nackErr := d.Nack(false, true); nackErr != nil {
									log.Printf("[queue:%s] Failed to nack message after panic: %v", queueName, nackErr)
								}
							}
						}()

						if err := c.handler(string(d.Body)); err != nil {
							log.Printf("[queue:%s] Error consuming: %s, %v", queueName, string(d.Body), err)
							if nackErr := d.Nack(false, true); nackErr != nil {
								log.Printf("[queue:%s] Failed to nack message: %v", queueName, nackErr)
							}
						} else {
							/*
								multiple：是否批量确认。
								false：仅确认当前 deliveryTag 对应的这一条消息。
								true：确认当前及之前所有未被确认的消息（即 deliveryTag ≤ 当前 tag 的消息）。
							*/
							if ackErr := d.Ack(false); ackErr != nil {
								log.Printf("[queue:%s] Failed to ack message: %v", queueName, ackErr)
							}
						}
					}()
				}
			}(que.Name, msgs)
		}
	}
	c.wg.Wait()
}

func (c *consumer) Stop() {
	c.channel.Close()
	c.conn.Close()
}

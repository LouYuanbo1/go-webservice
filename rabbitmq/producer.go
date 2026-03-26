package rabbitmq

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

type Producer interface {
	Produce(exchange string, routeKey string, msg []byte) error
}

type producer struct {
	conn        *amqp091.Connection
	channel     *amqp091.Channel
	contentType string
}

func NewProducer(producerConfig RabbitProducerConfig) (Producer, error) {
	producer := &producer{contentType: producerConfig.ContentType}
	conn, err := amqp091.Dial(getRabbitURL(producerConfig.RabbitConfig))
	if err != nil {
		return nil, err
	}

	producer.conn = conn
	channel, err := producer.conn.Channel()
	if err != nil {
		return nil, err
	}

	producer.channel = channel
	return producer, nil
}

func (p *producer) Produce(exchange string, routeKey string, msg []byte) error {
	return p.channel.PublishWithContext(
		context.Background(),
		exchange,
		routeKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: p.contentType,
			Body:        msg,
		},
	)
}

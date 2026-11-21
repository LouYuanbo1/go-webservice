package rabbitmq

import "errors"

var (
	ErrConnectFailed         = errors.New("event/rabbitmq: connect failed")
	ErrChannelFailed         = errors.New("event/rabbitmq: channel failed")
	ErrExchangeDeclareFailed = errors.New("event/rabbitmq: exchange declare failed")
	ErrQueueDeclareFailed    = errors.New("event/rabbitmq: queue declare failed")
	ErrQueueBindFailed       = errors.New("event/rabbitmq: queue bind failed")
	ErrConsumeFailed         = errors.New("event/rabbitmq: consume failed")
	ErrPublishFailed         = errors.New("event/rabbitmq: publish failed")
	ErrClosed                = errors.New("event/rabbitmq: bus closed")
)

package rabbitmq

import "errors"

var (
	ErrConnectFailed         = errors.New("rabbitmq: connect failed")
	ErrChannelFailed         = errors.New("rabbitmq: channel failed")
	ErrExchangeDeclareFailed = errors.New("rabbitmq: exchange declare failed")
	ErrQueueDeclareFailed    = errors.New("rabbitmq: queue declare failed")
	ErrQueueBindFailed       = errors.New("rabbitmq: queue bind failed")
	ErrProduceFailed         = errors.New("rabbitmq: produce failed")
	ErrConsumeFailed         = errors.New("rabbitmq: consume failed")
)

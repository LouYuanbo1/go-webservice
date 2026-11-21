package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/LouYuanbo1/go-webservice/event"
	amqp "github.com/rabbitmq/amqp091-go"
)

var _ event.Bus = (*Bus)(nil)

const defaultExchange = "event.bus"

// Bus 是基于 RabbitMQ topic 交换机的分布式事件总线，实现 event.Bus 接口。
// 每个订阅独占一个临时队列（autoDelete）并绑定到主题路由键，取消订阅即删除队列。
// 消息默认持久化到队列（durable 交换机），订阅者在线期间可靠投递。
type Bus struct {
	conn     *amqp.Connection
	pubCh    *amqp.Channel // 发布专用 channel，串行化保护，RabbitMQ channel 线程不安全
	pubMu    sync.Mutex
	exchange string

	mu     sync.Mutex
	subs   map[string]*subscription
	closed bool
}

// subscription 表示一条 RabbitMQ 订阅，独占一个消费 channel 与临时队列。
type subscription struct {
	ch      *amqp.Channel
	queue   string
	handler event.EventHandler
}

// NewBus 创建基于 RabbitMQ topic 交换机的事件总线，声明 topic 类型交换机并探测连接。
func NewBus(cfg *Config) (*Bus, error) {
	if cfg == nil {
		return nil, errorx.NewWithDetails(
			ErrConnectFailed,
			"event/rabbitmq",
			"NewBus",
			"rabbitmq config cannot be nil",
			nil,
		)
	}
	exchange := cfg.Exchange
	if exchange == "" {
		exchange = defaultExchange
	}

	conn, err := amqp.Dial(rabbitURL(*cfg))
	if err != nil {
		return nil, errorx.New(
			ErrConnectFailed,
			"event/rabbitmq",
			"NewBus",
			err,
		)
	}
	pubCh, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, errorx.New(
			ErrChannelFailed,
			"event/rabbitmq",
			"NewBus",
			err,
		)
	}
	if err := pubCh.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		_ = pubCh.Close()
		_ = conn.Close()
		return nil, errorx.New(
			ErrExchangeDeclareFailed,
			"event/rabbitmq",
			"NewBus",
			err,
		)
	}
	return &Bus{
		conn:     conn,
		pubCh:    pubCh,
		exchange: exchange,
		subs:     make(map[string]*subscription),
	}, nil
}

// rabbitURL 组装 amqp 连接串，VHost 为空时省略路径段。
func rabbitURL(cfg Config) string {
	vhost := cfg.VHost
	if vhost == "" || vhost == "/" {
		return fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, vhost)
}

// Subscribe 订阅主题：创建独占临时队列并绑定交换机路由键，然后消费投递。
// 同一主题当前只允许一个订阅者。
func (b *Bus) Subscribe(topic string, handler event.EventHandler) (event.Subscription, error) {
	if handler == nil {
		return nil, errorx.NewWithDetails(
			ErrConsumeFailed,
			"event/rabbitmq",
			"Subscribe",
			"handler is nil",
			nil,
		)
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil, errorx.New(
			ErrClosed,
			"event/rabbitmq",
			"Subscribe",
			nil,
		)
	}
	if _, dup := b.subs[topic]; dup {
		b.mu.Unlock()
		return nil, errorx.NewWithDetails(
			ErrConsumeFailed,
			"event/rabbitmq",
			"Subscribe",
			fmt.Sprintf("topic %q already subscribed", topic),
			nil,
		)
	}
	b.mu.Unlock()

	ch, err := b.conn.Channel()
	if err != nil {
		return nil, errorx.New(
			ErrChannelFailed,
			"event/rabbitmq",
			"Subscribe",
			err,
		)
	}
	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		_ = ch.Close()
		return nil, errorx.NewWithDetails(
			ErrQueueDeclareFailed,
			"event/rabbitmq",
			"Subscribe",
			fmt.Sprintf("declare queue for topic %q failed", topic),
			err,
		)
	}
	if err := ch.QueueBind(q.Name, topic, b.exchange, false, nil); err != nil {
		_ = ch.Close()
		return nil, errorx.NewWithDetails(
			ErrQueueBindFailed,
			"event/rabbitmq",
			"Subscribe",
			fmt.Sprintf("bind queue %q to topic %q failed", q.Name, topic),
			err,
		)
	}

	sub := &subscription{ch: ch, queue: q.Name, handler: handler}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		_ = ch.Close()
		return nil, errorx.New(
			ErrClosed,
			"event/rabbitmq",
			"Subscribe",
			nil,
		)
	}
	b.subs[topic] = sub
	b.mu.Unlock()

	go b.consume(topic, sub)

	var once sync.Once
	return &subscriptionHandle{once: &once, cl: func() { b.unsubscribe(topic, sub) }}, nil
}

// consume 持续读取投递并在手动确认模式（autoAck=false）下调用 handler。
// handler 返回错误时消息重新入队，避免丢失。
func (b *Bus) consume(topic string, sub *subscription) {
	msgs, err := sub.ch.Consume(sub.queue, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("[event/rabbitmq] consume topic %q queue %q: %v", topic, sub.queue, err)
		return
	}
	for d := range msgs {
		if err := sub.handler(context.Background(), d.Body); err != nil {
			log.Printf("[event/rabbitmq] handler error on topic %q: %v", topic, err)
			_ = d.Nack(false, true) // requeue
			continue
		}
		_ = d.Ack(false)
	}
}

// unsubscribe 幂等地取消订阅：取消消费者、删除临时队列并关闭 channel。
func (b *Bus) unsubscribe(topic string, sub *subscription) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	cur, ok := b.subs[topic]
	if !ok || cur != sub {
		b.mu.Unlock()
		return
	}
	delete(b.subs, topic)
	b.mu.Unlock()

	_ = sub.ch.Cancel(sub.queue, false)
	_, _ = sub.ch.QueueDelete(sub.queue, false, false, false)
	_ = sub.ch.Close()
}

// Publish 向主题发布一条携带字节负载的消息，路由键即主题。
func (b *Bus) Publish(ctx context.Context, topic string, payload []byte) error {
	b.mu.Lock()
	closed := b.closed
	b.mu.Unlock()
	if closed {
		return errorx.New(
			ErrClosed,
			"event/rabbitmq",
			"Publish",
			nil,
		)
	}

	b.pubMu.Lock()
	defer b.pubMu.Unlock()
	if err := b.pubCh.PublishWithContext(
		ctx,
		b.exchange,
		topic,
		false,
		false,
		amqp.Publishing{Body: payload},
	); err != nil {
		return errorx.NewWithDetails(
			ErrPublishFailed,
			"event/rabbitmq",
			"Publish",
			fmt.Sprintf("publish topic %q failed", topic),
			err,
		)
	}
	return nil
}

// Shutdown 关闭总线：逐个取消订阅、关闭发布 channel 与连接。幂等。
func (b *Bus) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	subs := make([]*subscription, 0, len(b.subs))
	for _, sub := range b.subs {
		subs = append(subs, sub)
	}
	b.subs = map[string]*subscription{}
	b.mu.Unlock()

	for _, sub := range subs {
		_ = sub.ch.Close()
	}
	_ = b.pubCh.Close()
	if err := b.conn.Close(); err != nil {
		return errorx.New(
			ErrClosed,
			"event/rabbitmq",
			"Shutdown",
			err,
		)
	}
	return nil
}

// subscriptionHandle 包装取消逻辑，保证 Unsubscribe 只执行一次。
type subscriptionHandle struct {
	once *sync.Once
	cl   func()
}

func (h *subscriptionHandle) Unsubscribe() {
	h.once.Do(h.cl)
}

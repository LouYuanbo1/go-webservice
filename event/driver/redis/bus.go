package redis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/LouYuanbo1/go-webservice/event"

	goredis "github.com/redis/go-redis/v9"
)

var _ event.Bus = (*Bus)(nil)

// Bus 是基于 Redis Pub/Sub 的分布式事件总线，实现 event.Bus 接口。
// 特点：消息为"发布即丢弃"，不持久化、支持多个订阅者并发投递。
type Bus struct {
	client *goredis.Client

	mu     sync.Mutex
	subs   map[string]*subscription // topic -> 订阅
	closed bool
}

// subscription 表示一条 Redis 订阅，同一连接只订阅单一 topic，便于独立取消。
type subscription struct {
	ps      *goredis.PubSub
	cancel  context.CancelFunc
	handler event.EventHandler
}

// NewBus 创建一个基于 Redis Pub/Sub 的事件总线，会即时探测连接可用性。
func NewBus(cfg *Config) (*Bus, error) {
	if cfg == nil {
		return nil, errorx.NewWithDetails(
			ErrConnectFailed,
			"event/redis",
			"NewBus",
			"redis config cannot be nil",
			nil,
		)
	}
	client := goredis.NewClient(&goredis.Options{
		Addr:        fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:    cfg.Password,
		DB:          cfg.DB,
		DialTimeout: 3 * time.Second,
		MaxRetries:  1,
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errorx.New(
			ErrConnectFailed,
			"event/redis",
			"NewBus",
			err,
		)
	}
	return &Bus{
		client: client,
		subs:   make(map[string]*subscription),
	}, nil
}

// Subscribe 订阅主题：为每个订阅建立独立 PubSub 连接并启动消费 goroutine。
func (b *Bus) Subscribe(topic string, handler event.EventHandler) (event.Subscription, error) {
	if handler == nil {
		return nil, errorx.NewWithDetails(
			ErrSubscribeFailed,
			"event/redis",
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
			"event/redis",
			"Subscribe",
			nil,
		)
	}
	if _, dup := b.subs[topic]; dup {
		b.mu.Unlock()
		return nil, errorx.NewWithDetails(
			ErrSubscribeFailed,
			"event/redis",
			"Subscribe",
			fmt.Sprintf("topic %q already subscribed", topic),
			nil,
		)
	}
	b.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	ps := b.client.Subscribe(ctx, topic)
	if _, err := ps.Receive(ctx); err != nil {
		cancel()
		return nil, errorx.NewWithDetails(
			ErrSubscribeFailed,
			"event/redis",
			"Subscribe",
			fmt.Sprintf("subscribe topic %q failed", topic),
			err,
		)
	}

	sub := &subscription{
		ps:      ps,
		cancel:  cancel,
		handler: handler,
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		cancel()
		_ = ps.Close()
		return nil, errorx.New(
			ErrClosed,
			"event/redis",
			"Subscribe",
			nil,
		)
	}
	b.subs[topic] = sub
	b.mu.Unlock()

	go b.consume(sub)

	var once sync.Once
	return &subscriptionHandle{once: &once, cl: func() { b.unsubscribe(topic, sub) }}, nil
}

// consume 持续读取 PubSub 消息并调用 handler。
func (b *Bus) consume(sub *subscription) {
	msgs := sub.ps.Channel()
	for msg := range msgs {
		if err := sub.handler(context.Background(), []byte(msg.Payload)); err != nil {
			log.Printf("[event/redis] handler error on topic %q: %v", msg.Channel, err)
		}
	}
}

// unsubscribe 幂等地取消订阅并释放连接，重复调用仅第一次生效。
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

	sub.cancel()
	_ = sub.ps.Close()
}

// Publish 向主题发布一条携带字节负载的消息。
func (b *Bus) Publish(ctx context.Context, topic string, payload []byte) error {
	b.mu.Lock()
	closed := b.closed
	b.mu.Unlock()
	if closed {
		return errorx.New(
			ErrClosed,
			"event/redis",
			"Publish",
			nil,
		)
	}
	if err := b.client.Publish(ctx, topic, payload).Err(); err != nil {
		return errorx.NewWithDetails(
			ErrPublishFailed,
			"event/redis",
			"Publish",
			fmt.Sprintf("publish topic %q failed", topic),
			err,
		)
	}
	return nil
}

// Shutdown 关闭总线：取消全部订阅、关闭 PubSub 与底层连接。幂等。
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
		sub.cancel()
		_ = sub.ps.Close()
	}
	if err := b.client.Close(); err != nil {
		return errorx.New(
			ErrClosed,
			"event/redis",
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

package event

import (
	"context"
	"fmt"
	"sync"
)

// LocalSub 表示一条本地订阅，调用 Unsubscribe 可取消订阅。
// 它同时满足 Subscription 接口，便于在需要时向下转型使用。
type LocalSub func()

// Unsubscribe 取消订阅。
func (s LocalSub) Unsubscribe() { s() }

// 本地订阅项，handler 以 any 形式持有，dispatch 时由泛型方法还原为具体结构体。
type localSubscription struct {
	fn func(ctx context.Context, event any) error
}

// LocalBus 是进程内事件总线，直接在内存中把结构体派发给订阅者，不做序列化。
// 它刻意不实现 Bus 接口（Bus 面向字节负载），因此可以直接使用泛型方法与结构体交互。
// 适用于单进程内的模块间解耦；跨进程请使用实现了 Bus 的分布式实现配合 TypedBus。
type LocalBus struct {
	mu sync.RWMutex
	// 主题 -> 订阅列表
	subs   map[string][]*localSubscription
	closed bool
}

// NewLocalBus 创建并返回一个本地事件总线。
func NewLocalBus() *LocalBus {
	return &LocalBus{subs: make(map[string][]*localSubscription)}
}

// Subscribe 订阅主题，handler 直接接收结构体事件 T，不做序列化。
// 未通过 WithTopic 指定主题时，默认使用基于类型名生成的主题。
func (b *LocalBus) Subscribe[T any](
	handler func(ctx context.Context, event T) error,
	opts ...Option,
) LocalSub {
	cfg := applyOptions(opts)
	topic := cfg.topic
	if topic == "" {
		topic = defaultTopic[T]()
	}

	fn := func(ctx context.Context, event any) error {
		typed, ok := event.(T)
		if !ok {
			return fmt.Errorf("localbus: event on %q is %T, want %T", topic, event, typed)
		}
		if cfg.filter != nil && !cfg.filter(typed) {
			return nil
		}
		return handler(ctx, typed)
	}

	sub := &localSubscription{fn: fn}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return func() {}
	}
	b.subs[topic] = append(b.subs[topic], sub)

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		list := b.subs[topic]
		for i, s := range list {
			if s == sub {
				b.subs[topic] = append(list[:i], list[i+1:]...)
				if len(b.subs[topic]) == 0 {
					delete(b.subs, topic)
				}
				return
			}
		}
	}
}

// Publish 将结构体事件 T 直接派发给同一主题下的所有订阅者。
// 未通过 WithTopic 指定主题时，默认使用基于类型名生成的主题。
func (b *LocalBus) Publish[T any](ctx context.Context, event T, opts ...Option) error {
	cfg := applyOptions(opts)
	topic := cfg.topic
	if topic == "" {
		topic = defaultTopic[T]()
	}

	b.mu.RLock()
	list := b.subs[topic]
	handlers := make([]func(context.Context, any) error, len(list))
	for i, s := range list {
		handlers[i] = s.fn
	}
	b.mu.RUnlock()

	for _, fn := range handlers {
		if err := fn(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown 关闭总线，清空所有订阅。已关闭的总线后续 Subscribe 与 Publish 均不生效。
func (b *LocalBus) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	b.subs = map[string][]*localSubscription{}
	return nil
}

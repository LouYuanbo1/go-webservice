package event

import (
	"context"
)

// EventHandler 处理某主题的事件，由 TypedBus 反序列化、过滤后间接调用。
type EventHandler func(ctx context.Context, payload []byte) error

// Subscription 表示一个已订阅的主题，调用 Unsubscribe 可取消订阅。
type Subscription interface {
	Unsubscribe()
}

// RawBus 保持不变，仍是传输层的非泛型接口
type Bus interface {
	Subscribe(topic string, handler EventHandler) (Subscription, error)
	Publish(ctx context.Context, topic string, payload []byte) error
	Shutdown(ctx context.Context) error
}

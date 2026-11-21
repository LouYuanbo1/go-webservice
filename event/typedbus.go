package event

import (
	"context"
	"fmt"
	"reflect"
)

// ---------- TypedBus ----------
// TypedBus 是泛型包装层，承载泛型方法
// T 是整个 Bus 实例的事件类型，也可以考虑不在类型上参数化，而是在方法上参数化
type TypedBus struct {
	raw   Bus
	codec Codec    // 实例级编解码器，默认 JSONCodec，不绑定具体类型
	opts  []Option // 订阅/发布级默认选项
}

func NewTypedBus(raw Bus, opts ...Option) *TypedBus {
	cfg := applyOptions(opts)
	codec := cfg.codec
	if codec == nil {
		codec = JSONCodec{}
	}
	return &TypedBus{raw: raw, codec: codec, opts: opts}
}

// defaultTopic 基于类型名生成默认主题。
func defaultTopic[T any]() string {
	var zero T
	name := reflect.TypeOf(zero).Name()
	if name == "" {
		name = fmt.Sprintf("%T", zero)
	}
	return "event." + name
}

// Subscribe 现在是 TypedBus 的方法，并可以声明自己的类型参数
// 注意：这里的 T 是方法级类型参数，与 TypedBus 本身无关
func (b *TypedBus) Subscribe[T any](
	handler func(ctx context.Context, event T) error,
	opts ...Option,
) (Subscription, error) {
	allOpts := append(b.opts, opts...)
	cfg := applyOptions(allOpts)
	topic := cfg.topic
	if topic == "" {
		topic = defaultTopic[T]()
	}

	rawHandler := func(ctx context.Context, payload []byte) error {
		var event T
		if err := b.codec.Unmarshal(payload, &event); err != nil {
			return fmt.Errorf("unmarshal %s: %w", topic, err)
		}
		if cfg.filter != nil && !cfg.filter(event) {
			return nil
		}
		return handler(ctx, event)
	}
	return b.raw.Subscribe(topic, rawHandler)
}

// Publish 同样是 TypedBus 的方法
func (b *TypedBus) Publish[T any](
	ctx context.Context,
	event T,
	opts ...Option,
) error {
	allOpts := append(b.opts, opts...)
	cfg := applyOptions(allOpts)
	topic := cfg.topic
	if topic == "" {
		topic = defaultTopic[T]()
	}
	payload, err := b.codec.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", topic, err)
	}
	return b.raw.Publish(ctx, topic, payload)
}

// Shutdown 透传关闭信号给底层 RawBus。
func (b *TypedBus) Shutdown(ctx context.Context) error {
	if b.raw == nil {
		return nil
	}
	return b.raw.Shutdown(ctx)
}

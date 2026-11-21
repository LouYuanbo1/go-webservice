package event

// ---------- 配置项 ----------

// Option 是 TypedBus 的可选配置项。
type Option func(o *options)

type options struct {
	topic  string
	filter func(any) bool
	codec  Codec
}

// applyOptions 依次应用所有 Option，nil 项会被安全跳过。
func applyOptions(all []Option) *options {
	opts := &options{}
	for _, opt := range all {
		if opt != nil {
			opt(opts)
		}
	}
	return opts
}

// WithTopic 设置事件的发布/订阅主题。
func WithTopic(topic string) Option {
	return func(o *options) { o.topic = topic }
}

// WithFilter 设置订阅过滤器，返回 false 的事件将被忽略。
func WithFilter[T any](fn func(event T) bool) Option {
	return func(o *options) {
		o.filter = func(v any) bool { return fn(v.(T)) }
	}
}

// WithCodec 设置编解码器，默认使用 JSONCodec，仅对 TypedBus 构造生效。
func WithCodec(c Codec) Option {
	return func(o *options) { o.codec = c }
}

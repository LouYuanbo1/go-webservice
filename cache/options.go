package cache

type Options struct {
	Prefix string
	// 只放真正通用的配置
}

type Option func(*Options)

func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}

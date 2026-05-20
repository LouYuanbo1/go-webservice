package breaker

import "time"

// ========== 配置选项 ==========

type options struct {
	k          float64       // Google SRE 算法中的 K 值（越大越敏感），默认 1.5
	protection int           // 保护请求数，确保至少放行这么多请求，默认 5
	window     time.Duration // 滑动窗口时长，默认 10s
	buckets    int           // 桶数量，默认 40
	name       string        // 熔断器名称（用于监控与日志）
	onReject   func()        // 拒绝回调（可用于打点/告警）
}

func defaultOptions() *options {
	return &options{
		k:          1.5,
		protection: 5,
		window:     10 * time.Second,
		buckets:    40,
		name:       "google-breaker",
		onReject:   nil,
	}
}

// Option 熔断器可选配置
type Option func(o *options)

// WithK 设置 Google SRE 算法中的 K 值（越大越敏感）
func WithK(k float64) Option {
	return func(o *options) { o.k = k }
}

// WithProtection 设置保护请求数，确保至少放行这么多请求
func WithProtection(p int) Option {
	return func(o *options) { o.protection = p }
}

// WithWindow 设置滑动窗口时长
func WithWindow(d time.Duration) Option {
	return func(o *options) { o.window = d }
}

// WithBuckets 设置桶数量
func WithBuckets(b int) Option {
	return func(o *options) { o.buckets = b }
}

// WithName 设置熔断器名称（用于监控与日志）
func WithName(name string) Option {
	return func(o *options) { o.name = name }
}

// WithRejectCallback 设置拒绝回调（可用于打点/告警）
func WithRejectCallback(fn func()) Option {
	return func(o *options) { o.onReject = fn }
}

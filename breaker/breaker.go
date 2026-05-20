package breaker

import (
	"context"
	"errors"
	"fmt"
	"math"
)

// ========== 熔断器接口 ==========

// Breaker 熔断器接口，与 Go-zero 风格一致
type Breaker interface {
	// Do 执行请求，若返回 nil 则视为成功
	Do(ctx context.Context, req func(ctx context.Context) error) error
	// DoWithAcceptable 执行请求，用 acceptable 自定义成功判定
	DoWithAcceptable(ctx context.Context, req func(ctx context.Context) error, acceptable func(err error) bool) error
	// DoWithFallback 执行请求，支持降级函数
	DoWithFallback(ctx context.Context, req func(ctx context.Context) error, fallback func(err error) error) error
	// GetMetrics 获取当前窗口指标（用于 Prometheus 导出）
	GetMetrics() (total, accepts int64, rate float64)
	// Reset 手动重置（运维干预用）
	Reset()
}

// ========== Google SRE 自适应熔断器实现 ==========

type googleBreaker struct {
	k          float64
	protection float64
	stat       *rollingWindow
	proba      *proba
	name       string
	onReject   func()
}

// NewBreaker 创建一个 Google SRE 自适应熔断器
func NewBreaker(opts ...Option) Breaker {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	return &googleBreaker{
		k:          o.k,
		protection: float64(o.protection),
		stat:       newRollingWindow(o.window, o.buckets),
		proba:      newProba(),
		name:       o.name,
		onReject:   o.onReject,
	}
}

// Do 执行请求，支持 Context
func (b *googleBreaker) Do(ctx context.Context, req func(ctx context.Context) error) error {
	return b.DoWithFallback(ctx, req, nil)
}

// DoWithAcceptable 执行请求，由 acceptable 决定是否标记为成功
func (b *googleBreaker) DoWithAcceptable(ctx context.Context, req func(ctx context.Context) error, acceptable func(err error) bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := b.accept(); err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			b.stat.add(false)
			panic(r)
		}
	}()

	err := req(ctx)
	b.stat.add(acceptable(err))
	return err
}

// DoWithFallback 执行请求，支持 Context 和降级函数
func (b *googleBreaker) DoWithFallback(ctx context.Context, req func(ctx context.Context) error, fallback func(err error) error) error {
	if err := ctx.Err(); err != nil {
		if fallback != nil {
			return fallback(err)
		}
		return err
	}

	if err := b.accept(); err != nil {
		if fallback != nil {
			return fallback(err)
		}
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			b.stat.add(false)
			if fallback != nil {
				fallback(fmt.Errorf("panic: %v", r))
			}
		}
	}()

	err := req(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			b.stat.add(false)
		} else {
			b.stat.add(false)
		}
		if fallback != nil {
			return fallback(err)
		}
		return err
	}
	b.stat.add(true)
	return nil
}

// GetMetrics 获取当前窗口指标（用于 Prometheus 导出）
func (b *googleBreaker) GetMetrics() (total, accepts int64, rate float64) {
	accepts, total = b.stat.history()
	if total > 0 {
		rate = float64(accepts) / float64(total)
	}
	return
}

// Reset 手动重置（运维干预用）
func (b *googleBreaker) Reset() {
	b.stat.reset()
}

// accept 根据滑动窗口历史数据计算拒绝概率并试探
func (b *googleBreaker) accept() error {
	accepts, total := b.stat.history()
	weightedAccepts := b.k * float64(accepts)
	// Google SRE 公式：dropRatio = max(0, (total - protection - k*accepts) / (total + 1))
	dropRatio := math.Max(0, (float64(total)-b.protection-weightedAccepts)/float64(total+1))
	if dropRatio <= 0 {
		return nil
	}
	if b.proba.TrueOnProba(dropRatio) {
		if b.onReject != nil {
			go b.onReject()
		}
		return ErrServiceUnavailable
	}
	return nil
}

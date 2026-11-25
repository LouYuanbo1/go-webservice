// Package antx 提供基于 ant 原生 Interceptor 中间件的熔断保护。
// 用法：在 ant.Open 时通过 WithBreaker 安装，即可为该 Context 及其
// 派生的全部 Session 的所有 SQL 语句提供熔断保护。
package antx

import (
	"context"
	"database/sql"

	"github.com/LouYuanbo1/go-webservice/breaker"
	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/skudra-go/ant"
)

// DefaultAcceptable 判断一条已执行语句的错误是否视为可容忍（即熔断器窗口记成功）。
// 当前容忍：nil（执行成功）、sql.ErrNoRows（查无记录）、sql.ErrTxDone（事务已结束）。
func DefaultAcceptable(err error) bool {
	if err == nil {
		return true
	}
	return errorx.In(err, sql.ErrNoRows, sql.ErrTxDone)
}

// WithBreaker 为整个 Context 安装熔断保护（基于 ant 原生 Interceptor 中间件）。
//
// 熔断器实例在调用 ant.Open 时创建一次，被该 Context 派生的所有 Session
// 共享，失败统计跨请求正常累积；它会拦截该上下文执行的每一条 SQL
// （含 Table 方法、Query 链、SaveChanges、Raw、事务），熔断时在触库前
// 直接拒绝并返回 breaker.ErrServiceUnavailable。
//
// opts 用于定制熔断算法（WithK、WithProtection 等）与拒绝回调
// （WithRejectCallback），不传则使用默认配置。用法：
//
//	db, _ := ant.Open[AppDB](ctx, sqlite.Dialect, dsn, antx.WithBreaker())
func WithBreaker(opts ...breaker.Option) ant.Option {
	return WithCustomBreaker(breaker.NewGoogleBreaker(opts...), DefaultAcceptable)
}

// WithCustomBreaker 使用自建熔断器与自定义成功判定函数安装熔断中间件，
// 便于复用外部 Breaker 实现、接入监控打点或覆盖 acceptable 逻辑。
func WithCustomBreaker(brk breaker.Breaker, acceptable func(error) bool) ant.Option {
	if acceptable == nil {
		acceptable = DefaultAcceptable
	}
	return ant.WithInterceptor(&breakerInterceptor{brk: brk, acceptable: acceptable})
}

type breakerInterceptor struct {
	brk        breaker.Breaker
	acceptable func(error) bool
}

var _ ant.Interceptor = (*breakerInterceptor)(nil)

// Before 在语句触库前做准入判定：上下文已失效或熔断拒绝时直接中断语句。
func (i *breakerInterceptor) Before(ctx context.Context, _ *ant.Statement) (context.Context, error) {
	if err := ctx.Err(); err != nil {
		return ctx, err
	}
	if err := i.brk.Allow(); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// After 按语句的真实结果回写熔断窗口：可容忍的错误记成功，其余记失败。
func (i *breakerInterceptor) After(_ context.Context, _ *ant.Statement, o ant.Outcome) {
	if i.acceptable(o.Err) {
		i.brk.MarkSuccess()
	} else {
		i.brk.MarkFailure()
	}
}

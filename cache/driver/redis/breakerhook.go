package redis

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/breaker"
	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/redis/go-redis/v9"
)

var ignoreCmds = map[string]struct{}{
	"blpop": {},
}

func acceptable(err error) bool {
	return err == nil || errorx.In(err, redis.Nil, context.Canceled)
}

type BreakerHook struct {
	brk breaker.Breaker
}

func NewBreakerHook(brk breaker.Breaker) BreakerHook {
	return BreakerHook{brk: brk}
}

func (h BreakerHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h BreakerHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if _, ok := ignoreCmds[cmd.Name()]; ok {
			return next(ctx, cmd)
		}

		return h.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
			return next(ctx, cmd)
		}, acceptable)
	}
}

func (h BreakerHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return h.brk.DoWithAcceptable(ctx, func(ctx context.Context) error {
			return next(ctx, cmds)
		}, acceptable)
	}
}

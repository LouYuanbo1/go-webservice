package gen

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/gormc"
	"github.com/LouYuanbo1/go-webservice/gormx/gen"
)

type ExecFn[T any, ID comparable, PT gen.PointerModel[T, ID]] func(ctx context.Context, conn gen.Session[T, ID, PT]) error
type QueryFn[T any, ID comparable, PT gen.PointerModel[T, ID]] func(ctx context.Context, conn gen.Session[T, ID, PT], val PT) error
type PrimaryQueryFn[T any, ID comparable, PT gen.PointerModel[T, ID]] func(ctx context.Context, conn gen.Session[T, ID, PT], val PT, primaryKey ID) error

type CachedConn[T any, ID comparable, PT gen.PointerModel[T, ID]] interface {
	GetCache(ctx context.Context, key string, val PT) error
	DelCache(ctx context.Context, key ...string) error
	Exec(ctx context.Context, exec ExecFn[T, ID, PT], keys ...string) error
	Query(ctx context.Context, val PT, key string, query QueryFn[T, ID, PT], opts ...gormc.TTLOption) error
	QueryIndex(ctx context.Context, val PT, key string, keyer func(primary ID) string, indexQuery QueryFn[T, ID, PT], primaryQuery PrimaryQueryFn[T, ID, PT], opts ...gormc.TTLOption) error
}

type cachedConn[T any, ID comparable, PT gen.PointerModel[T, ID]] struct {
	conn  gen.Session[T, ID, PT]
	cache cache.Client
	cfg   *gormc.Config
}

func NewConnWithCache[T any, ID comparable, PT gen.PointerModel[T, ID]](db gen.Session[T, ID, PT], c cache.Client, cfg *gormc.Config) CachedConn[T, ID, PT] {
	return &cachedConn[T, ID, PT]{
		conn:  db,
		cache: c,
		cfg:   cfg,
	}
}

func (c *cachedConn[T, ID, PT]) GetCache(ctx context.Context, key string, val PT) error {
	return c.cache.Get(ctx, key, val)
}

func (c *cachedConn[T, ID, PT]) DelCache(ctx context.Context, key ...string) error {
	return c.cache.Del(ctx, key...)
}

// Cache-Aside
func (cc *cachedConn[T, ID, PT]) Exec(ctx context.Context, exec ExecFn[T, ID, PT], keys ...string) error {
	err := exec(ctx, cc.conn)
	if err != nil {
		return err
	}
	return cc.cache.Del(ctx, keys...)
}

func (cc *cachedConn[T, ID, PT]) Query(ctx context.Context, val PT, key string, query QueryFn[T, ID, PT], opts ...gormc.TTLOption) error {
	return cc.cache.Take(ctx, val, key, func(val any) error {
		return query(ctx, cc.conn, val.(PT))
	}, gormc.TTLBuilder(cc.cfg.TTL, opts...).GetTTL())
}

// 可能需要调整过期时间的关系避免缓存击穿或者雪崩
func (cc *cachedConn[T, ID, PT]) QueryIndex(ctx context.Context, val PT, key string, keyer func(primary ID) string, indexQuery QueryFn[T, ID, PT], primaryQuery PrimaryQueryFn[T, ID, PT], opts ...gormc.TTLOption) error {
	var primaryKey any
	//查询设置对应的主键
	if err := cc.cache.Take(ctx, primaryKey, key,
		func(val any) error {
			err := indexQuery(ctx, cc.conn, val.(PT))
			if err != nil {
				return err
			}
			return nil
		},
		gormc.TTLBuilder(cc.cfg.TTL, opts...).GetTTL()+cc.cfg.CacheSafeGapBetweenIndexAndPrimary,
	); err != nil {
		return err
	}

	if primaryKey == nil {
		return nil
	}

	//查询主键对应的值
	return cc.cache.Take(ctx, val, keyer(primaryKey.(ID)), func(val any) error {
		return primaryQuery(ctx, cc.conn, val.(PT), primaryKey.(ID))
	}, gormc.TTLBuilder(cc.cfg.TTL, opts...).GetTTL())
}

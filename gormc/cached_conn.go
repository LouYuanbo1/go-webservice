package gormc

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/gormx"
)

type ExecFn func(ctx context.Context, conn gormx.Conn) error
type QueryFn func(ctx context.Context, conn gormx.Conn, val any) error
type PrimaryQueryFn func(ctx context.Context, conn gormx.Conn, primaryKey, val any) error

type CachedConn interface {
	GetCache(ctx context.Context, key string, val any) error
	DelCache(ctx context.Context, key ...string) error
	Exec(ctx context.Context, exec ExecFn, keys ...string) error
	Query(ctx context.Context, key string, val any, query QueryFn, opts ...TTLOption) error
	QueryIndex(ctx context.Context, key string, val any, keyer func(primary any) string, indexQuery QueryFn, primaryQuery PrimaryQueryFn, opts ...TTLOption) error
}

type cachedConn struct {
	conn  gormx.Conn
	cache cache.Client
	cfg   *Config
}

func NewConnWithCache(db gormx.Conn, c cache.Client, cfg *Config) CachedConn {
	return &cachedConn{
		conn:  db,
		cache: c,
		cfg:   cfg,
	}
}

func (c *cachedConn) GetCache(ctx context.Context, key string, val any) error {
	return c.cache.Get(ctx, key, val)
}

func (c *cachedConn) DelCache(ctx context.Context, key ...string) error {
	return c.cache.Del(ctx, key...)
}

// Cache-Aside
func (cc *cachedConn) Exec(ctx context.Context, exec ExecFn, keys ...string) error {
	err := exec(ctx, cc.conn)
	if err != nil {
		return err
	}
	return cc.cache.Del(ctx, keys...)
}

func (cc *cachedConn) Query(ctx context.Context, key string, val any, query QueryFn, opts ...TTLOption) error {
	return cc.cache.Take(ctx, key, val, func(val any) error {
		return query(ctx, cc.conn, val)
	}, TTLBuilder(cc.cfg.TTL, opts...).GetTTL())
}

// 可能需要调整过期时间的关系避免缓存击穿或者雪崩
func (cc *cachedConn) QueryIndex(ctx context.Context, key string, val any, keyer func(primary any) string, indexQuery QueryFn, primaryQuery PrimaryQueryFn, opts ...TTLOption) error {
	var primaryKey any

	//查询设置对应的主键
	if err := cc.cache.Take(ctx, key, primaryKey,
		func(val any) (err error) {
			err = indexQuery(ctx, cc.conn, val)
			if err != nil {
				return err
			}
			return nil
		},
		TTLBuilder(cc.cfg.TTL, opts...).GetTTL() + cc.cfg.CacheSafeGapBetweenIndexAndPrimary,
	); err != nil {
		return err
	}

	if primaryKey == nil {
		return nil
	}

	//查询主键对应的值
	return cc.cache.Take(ctx, keyer(primaryKey), val, func(val any) error {
		return primaryQuery(ctx, cc.conn, primaryKey, val)
	}, TTLBuilder(cc.cfg.TTL, opts...).GetTTL())
}

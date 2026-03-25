package gormc

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/gormx"
)

type ExecFn func(ctx context.Context, conn gormx.Conn) error
type QueryFn func(ctx context.Context, conn gormx.Conn, val any) error
type IndexQueryFn func(ctx context.Context, conn gormx.Conn, val any) (primaryKey any, err error)
type PrimaryQueryFn func(ctx context.Context, conn gormx.Conn, val, primaryKey any) error

type CachedConn interface {
	GetCache(ctx context.Context, key string, val any) error
	DelCache(ctx context.Context, key ...string) error
	Exec(ctx context.Context, exec ExecFn, keys ...string) error
	Query(
		ctx context.Context,
		val any,
		key string,
		query QueryFn,
		opts ...TTLOption,
	) error
	QueryIndex(
		ctx context.Context,
		val any,
		key string,
		keyer func(primary any) string,
		indexQuery IndexQueryFn,
		primaryQuery PrimaryQueryFn,
		opts ...TTLOption,
	) error
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

func (cc *cachedConn) GetCache(ctx context.Context, key string, val any) error {
	return cc.cache.Get(ctx, key, val)
}

func (cc *cachedConn) DelCache(ctx context.Context, key ...string) error {
	return cc.cache.Del(ctx, key...)
}

// Cache-Aside
func (cc *cachedConn) Exec(ctx context.Context, exec ExecFn, keys ...string) error {
	err := exec(ctx, cc.conn)
	if err != nil {
		return err
	}
	return cc.cache.Del(ctx, keys...)
}

func (cc *cachedConn) Query(
	ctx context.Context,
	val any,
	key string,
	query QueryFn,
	opts ...TTLOption,
) error {
	return cc.cache.Take(ctx, val, key, func(cachedVal any) error {
		return query(ctx, cc.conn, cachedVal)
	}, cc.ttlBuilder(opts...).value)
}

// 可能需要调整过期时间的关系避免缓存击穿或者雪崩
func (cc *cachedConn) QueryIndex(
	ctx context.Context,
	val any,
	key string,
	keyer func(primary any) string,
	indexQuery IndexQueryFn,
	primaryQuery PrimaryQueryFn,
	opts ...TTLOption,
) error {
	var primaryKey any
	var foundPrimaryKeyFromDB bool
	/*
		如果缓存中有主键,通过&primaryKey获取,否则从数据库查询主键,并缓存主键对应的值
	*/
	if err := cc.cache.Take(ctx, &primaryKey, key,
		func(cachedVal any) (err error) {
			//如果缓存未命中,则从数据库查询主键,注意此时同时也已经给value赋值了
			pk, err:= indexQuery(ctx, cc.conn, val)
			if err != nil {
				return err
			}
			primaryKey = pk
			//将主键赋值给val,之后Take中的Set会自动缓存键对应的主键
			*cachedVal.(*any) = primaryKey
			foundPrimaryKeyFromDB = true
			/*
				缓存主键对应的值,过期时间为缓存安全间隔
			*/
			return cc.cache.Set(ctx, keyer(primaryKey), val, cc.ttlBuilder(opts...).value+cc.cfg.CacheSafeGapBetweenIndexAndPrimary)
		},
		cc.ttlBuilder(opts...).value,
	); err != nil {
		return err
	}

	/*
		如果 foundPrimaryKeyFromDB == true，说明索引缓存未命中，并且我们在回调中已经通过 indexQuery 把数据填充到 v 并写入了主键缓存。
		此时 v 已经包含了正确的结果，所以直接返回 nil。
		如果 foundPrimaryKeyFromDB == false，说明索引缓存命中，我们直接从缓存拿到了 primaryKey，此时 v 仍然是零值，还没有数据。
		因此需要继续通过主键缓存获取数据：
	*/
	if foundPrimaryKeyFromDB {
		return nil
	}

	//查询主键对应的值
	return cc.cache.Take(ctx, val, keyer(primaryKey), func(val any) error {
		return primaryQuery(ctx, cc.conn, val, primaryKey)
	}, cc.ttlBuilder(opts...).value)
}

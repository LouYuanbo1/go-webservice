package gormc

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/gormx"
)

type ExecFn func(ctx context.Context, db gormx.DB) error
type QueryFn func(ctx context.Context, db gormx.DB, val any) error
type IndexQueryFn func(ctx context.Context, db gormx.DB, val any) (primaryKey any, err error)
type PrimaryQueryFn func(ctx context.Context, db gormx.DB, val, primaryKey any) error

type CacheDB interface {
	GetCache(ctx context.Context, key string, val any) error
	SetCache(ctx context.Context, key string, val any, opts ...TTLOption) error
	DelCache(ctx context.Context, key ...string) error
	Exec(ctx context.Context, exec ExecFn, keys ...string) error
	Query(
		ctx context.Context,
		key string,
		val any,
		query QueryFn,
		opts ...TTLOption,
	) error
	QueryIndex(
		ctx context.Context,
		key string,
		val any,
		keyer func(primary any) string,
		indexQuery IndexQueryFn,
		primaryQuery PrimaryQueryFn,
		opts ...TTLOption,
	) error
	ExecNoCache(ctx context.Context, exec ExecFn) error
	QueryNoCache(ctx context.Context, val any, query QueryFn) error
	Transaction(ctx context.Context, fn func(ctx context.Context, sess gormx.Session) error) error
}

type cacheDB struct {
	db    gormx.DB
	cache cache.Client
	cfg   *Config
}

func NewDBWithCache(db gormx.DB, c cache.Client, cfg *Config) CacheDB {
	return &cacheDB{
		db:    db,
		cache: c,
		cfg:   cfg,
	}
}

func (cdb *cacheDB) GetCache(ctx context.Context, key string, val any) error {
	return cdb.cache.Get(ctx, key, val)
}

func (cdb *cacheDB) SetCache(ctx context.Context, key string, val any, opts ...TTLOption) error {
	return cdb.cache.Set(ctx, key, val, cdb.ttlBuilder(opts...).value)
}

func (cdb *cacheDB) DelCache(ctx context.Context, key ...string) error {
	return cdb.cache.Del(ctx, key...)
}

// Cache-Aside
func (cdb *cacheDB) Exec(ctx context.Context, exec ExecFn, keys ...string) error {
	err := exec(ctx, cdb.db)
	if err != nil {
		return err
	}
	return cdb.cache.Del(ctx, keys...)
}

func (cdb *cacheDB) Query(
	ctx context.Context,
	key string,
	val any,
	query QueryFn,
	opts ...TTLOption,
) error {
	return cdb.cache.Take(ctx, key, val, func(cachedVal any) error {
		return query(ctx, cdb.db, cachedVal)
	}, cdb.ttlBuilder(opts...).value)
}

// 可能需要调整过期时间的关系避免缓存击穿或者雪崩
func (cdb *cacheDB) QueryIndex(
	ctx context.Context,
	key string,
	val any,
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
	if err := cdb.cache.Take(ctx, key, &primaryKey,
		func(cachedVal any) (err error) {
			//如果缓存未命中,则从数据库查询主键,注意此时同时也已经给value赋值了
			pk, err := indexQuery(ctx, cdb.db, val)
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
			return cdb.cache.Set(ctx, keyer(primaryKey), val, cdb.ttlBuilder(opts...).value+cdb.cfg.CacheSafeGapBetweenIndexAndPrimary)
		},
		cdb.ttlBuilder(opts...).value,
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
	return cdb.cache.Take(ctx, keyer(primaryKey), val, func(val any) error {
		return primaryQuery(ctx, cdb.db, val, primaryKey)
	}, cdb.ttlBuilder(opts...).value)
}

func (cdb *cacheDB) ExecNoCache(ctx context.Context, exec ExecFn) error {
	return exec(ctx, cdb.db)
}

func (cdb *cacheDB) QueryNoCache(ctx context.Context, val any, query QueryFn) error {
	return query(ctx, cdb.db, val)
}

// 不建议在事务中使用缓存，因为事务中的缓存是不同的，会导致缓存不一致
func (cdb *cacheDB) Transaction(ctx context.Context, fn func(ctx context.Context, s gormx.Session) error) error {
	return cdb.db.Transaction(ctx, fn)
}

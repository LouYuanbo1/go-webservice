package gormc

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/gormx"
)

type CacheDB struct {
	db    *gormx.DB
	cache *cache.Client
	cfg   *Config
}

func NewCacheDB(db *gormx.DB, c *cache.Client, cfg *Config) *CacheDB {
	return &CacheDB{
		db:    db,
		cache: c,
		cfg:   cfg,
	}
}

func (cdb *CacheDB) GetXDB() *gormx.DB {
	return cdb.db
}

func (cdb *CacheDB) GetCache[T any](ctx context.Context, key string, val *T) error {
	return cdb.cache.Get(ctx, key, val)
}

func (cdb *CacheDB) SetCache(ctx context.Context, key string, val any, opts ...TTLOption) error {
	return cdb.cache.Set(ctx, key, val, cdb.ttlBuilder(opts...).value)
}

func (cdb *CacheDB) DelCache(ctx context.Context, key ...string) error {
	return cdb.cache.Del(ctx, key...)
}

func (cdb *CacheDB) Exec(ctx context.Context, exec func(ctx context.Context, db *gormx.DB) error, keys ...string) error {
	err := exec(ctx, cdb.db)
	if err != nil {
		return err
	}
	return cdb.cache.Del(ctx, keys...)
}

func (cdb *CacheDB) ExecNoCache(ctx context.Context, exec func(ctx context.Context, db *gormx.DB) error) error {
	return exec(ctx, cdb.db)
}

func (cdb *CacheDB) Query[T any](
	ctx context.Context,
	key string,
	val *T,
	query func(ctx context.Context, db *gormx.DB, val *T) error,
	opts ...TTLOption,
) error {
	return cdb.cache.Take(ctx, key, val, func(cachedVal *T) error {
		return query(ctx, cdb.db, cachedVal)
	}, cdb.ttlBuilder(opts...).value)
}

func (cdb *CacheDB) QueryIndex[T any, ID comparable](
	ctx context.Context,
	key string, // 索引缓存键
	val *T, // 最终数据的存放指针
	keyer func(primary ID) string, // 根据主键生成主键缓存键
	indexQuery func(ctx context.Context, db *gormx.DB, val *T) (primaryKey ID, err error), // 通过索引查主键
	primaryQuery func(ctx context.Context, db *gormx.DB, val *T, primaryKey ID) error, // 通过主键查数据
	opts ...TTLOption,
) error {
	var primaryKey ID
	var foundPrimaryKeyFromDB bool // 标记是否从数据库直接获取了数据

	// 1. 尝试获取索引缓存（存储主键）
	if err := cdb.cache.Take(ctx, key, &primaryKey,
		func(cachedVal *ID) error { // cachedVal 实际上是 *ID
			// 从数据库通过索引获取主键，并填充完整数据到 val
			pk, err := indexQuery(ctx, cdb.db, val)
			if err != nil {
				return err
			}
			primaryKey = pk
			// 将主键写入索引缓存（让 Take 自动写入）
			// cachedVal 类型为 *ID，需解引用赋值
			*cachedVal = primaryKey
			foundPrimaryKeyFromDB = true
			// 手动将完整数据写入主键缓存，过期时间略长于索引缓存
			return cdb.cache.Set(ctx, keyer(primaryKey), val,
				cdb.ttlBuilder(opts...).value+cdb.cfg.CacheSafeGapBetweenIndexAndPrimary)
		},
		cdb.ttlBuilder(opts...).value,
	); err != nil {
		return err
	}

	// 2. 如果已经从数据库获取了数据（索引缓存未命中），直接返回
	if foundPrimaryKeyFromDB {
		return nil
	}

	// 3. 索引缓存命中，得到 primaryKey，现在通过主键缓存获取完整数据
	return cdb.cache.Take(ctx, keyer(primaryKey), val, func(v *T) error {
		// 主键缓存未命中时，通过数据库查询并填充
		return primaryQuery(ctx, cdb.db, v, primaryKey)
	}, cdb.ttlBuilder(opts...).value)
}

func (cdb *CacheDB) QueryNoCache[T any](ctx context.Context, val *T, query func(ctx context.Context, db *gormx.DB, val *T) error) error {
	return query(ctx, cdb.db, val)
}

func (cdb *CacheDB) QueryRowsNoCache[T any](ctx context.Context, val *[]T, query func(ctx context.Context, db *gormx.DB, val *[]T) error) error {
	return query(ctx, cdb.db, val)
}

func (cdb *CacheDB) Transaction(ctx context.Context, fn func(tx *gormx.Executor) error) error {
	return cdb.db.Transaction(ctx, fn)
}

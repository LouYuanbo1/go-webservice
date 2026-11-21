package antc

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/cachedb"
	"github.com/skudra-go/ant"
)

// HasAntContext 是整库层对用户自定义数据库结构体指针的约束接口。
// ant 的 Tx 等能力是定义在 *ant.Context 上的泛型方法，普通接口无法捕获，
// 因此要求用户结构体实现 AntContext 方法返回其内嵌的 *ant.Context，
// 从而获得编译期安全的整库能力（事务等）。
//
// 与 PointerModel[T] 同款约束：类型集合限定为 *T，编译器可验证 T 的
// 指针类型确实可用。
//
// 用法示例（牺牲一点侵入性换取安全性）：
//
//	type AppDB struct {
//		ant.Context
//		Users ant.Table[User]
//	}
//
//	func (db *AppDB) AntContext() *ant.Context { return &db.Context }
type HasAntContext[T any] interface {
	*T
	AntContext() *ant.Context
}

// CacheDB 是整库缓存层：包住用户自定义的 DB 结构体（嵌入 ant.Context），
// 查询与写入闭包可直接访问任意表，实现跨表操作与统一缓存失效。
// 底层持有 *DB（类型参数 DB 即 *AppDB），事务通过 AntContext().Tx 转发给 ant。
type CacheDB[T any, DB HasAntContext[T]] struct {
	db    DB
	cache *cache.Client
	cfg   *cachedb.Config
}

// NewCacheDB 构造一个带缓存能力的整库包装器。
// db 为已由 ant.Open/Wrap 绑定返回的 *AppDB。
func NewCacheDB[T any, DB HasAntContext[T]](db DB, c *cache.Client, cfg *cachedb.Config) *CacheDB[T, DB] {
	return &CacheDB[T, DB]{
		db:    db,
		cache: c,
		cfg:   cfg,
	}
}

// GetDB 返回底层 *DB，便于执行未缓存的原生读写（含 ant.Context 提供的方法）。
func (cdb *CacheDB[T, DB]) GetDB() DB {
	return cdb.db
}

// GetCache 直接读取缓存键。
func (cdb *CacheDB[T, DB]) GetCache[V any](ctx context.Context, key string, val *V) error {
	return cdb.cache.Get(ctx, key, val)
}

// SetCache 直接写入缓存键，TTL 缺省使用配置默认值。
func (cdb *CacheDB[T, DB]) SetCache(ctx context.Context, key string, val any, opts ...cachedb.TTLOption) error {
	return cdb.cache.Set(ctx, key, val, cachedb.BuildTTL(cdb.cfg, opts...))
}

// DelCache 删除缓存键。
func (cdb *CacheDB[T, DB]) DelCache(ctx context.Context, key ...string) error {
	return cdb.cache.Del(ctx, key...)
}

// Exec 执行写入并删除相关缓存键，保证缓存与数据库一致。exec 可跨任意表写入。
func (cdb *CacheDB[T, DB]) Exec(ctx context.Context, exec func(ctx context.Context, db DB) error, keys ...string) error {
	err := exec(ctx, cdb.db)
	if err != nil {
		return err
	}
	return cdb.cache.Del(ctx, keys...)
}

// ExecNoCache 执行写入但不做任何缓存清理。
func (cdb *CacheDB[T, DB]) ExecNoCache(ctx context.Context, exec func(ctx context.Context, db DB) error) error {
	return exec(ctx, cdb.db)
}

// Query 缓存优先的单行查询：缓存命中直接反序列化返回，未命中执行
// query 回源数据库，成功后回填缓存。返回 (*R, error)，查询无结果时
// query 应返回 ant.ErrNotFound。
func (cdb *CacheDB[T, DB]) Query[R any](
	ctx context.Context,
	key string,
	query func(ctx context.Context, db DB) (*R, error),
	opts ...cachedb.TTLOption,
) (*R, error) {
	var result R
	err := cdb.cache.Take(ctx, key, &result, func(cachedVal *R) error {
		u, err := query(ctx, cdb.db)
		if err != nil {
			return err
		}
		if u == nil {
			return ant.ErrNotFound
		}
		*cachedVal = *u
		return nil
	}, cachedb.BuildTTL(cdb.cfg, opts...))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryIndex 两级缓存查询：先用索引缓存拿到主键，再按主键取完整数据。
// 索引缓存与主键缓存取不同 TTL，主键缓存稍长，避免索引缓存未过期而
// 主键缓存先行过期造成一次额外的数据库访问。
// indexQuery 通过索引定位主键并返回该行数据（用于回填主键缓存，避免二次查询）。
func (cdb *CacheDB[T, DB]) QueryIndex[R any, ID comparable](
	ctx context.Context,
	key string, // 索引缓存键（缓存内容为主键）
	keyer func(primary ID) string, // 根据主键生成主键缓存键
	indexQuery func(ctx context.Context, db DB) (primaryKey ID, data *R, err error), // 通过索引查主键并返回该行
	primaryQuery func(ctx context.Context, db DB, primaryKey ID) (*R, error), // 通过主键查数据
	opts ...cachedb.TTLOption,
) (*R, error) {
	ttl := cachedb.BuildTTL(cdb.cfg, opts...)

	var primaryKey ID
	var foundPrimaryKeyFromDB bool // 标记是否从数据库直接获取了数据
	var result R

	// 1. 尝试获取索引缓存（存储主键）
	if err := cdb.cache.Take(ctx, key, &primaryKey,
		func(cachedVal *ID) error {
			// 从数据库通过索引获取主键和该行数据
			pk, u, err := indexQuery(ctx, cdb.db)
			if err != nil {
				return err
			}
			if u == nil {
				return ant.ErrNotFound
			}
			*cachedVal = pk
			result = *u
			foundPrimaryKeyFromDB = true
			// 手动将完整数据写入主键缓存，过期时间略长于索引缓存
			return cdb.cache.Set(ctx, keyer(pk), *u,
				ttl+cachedb.SafeGap(cdb.cfg))
		},
		ttl,
	); err != nil {
		return nil, err
	}

	// 2. 如果已经从数据库获取了数据（索引缓存未命中），直接返回
	if foundPrimaryKeyFromDB {
		return &result, nil
	}

	// 3. 索引缓存命中，得到 primaryKey，现在通过主键缓存获取完整数据
	var v R
	if err := cdb.cache.Take(ctx, keyer(primaryKey), &v, func(cachedVal *R) error {
		u, err := primaryQuery(ctx, cdb.db, primaryKey)
		if err != nil {
			return err
		}
		if u == nil {
			return ant.ErrNotFound
		}
		*cachedVal = *u
		return nil
	}, ttl); err != nil {
		return nil, err
	}
	return &v, nil
}

// QueryNoCache 直接查询数据库，不经过缓存（也不回填）。
func (cdb *CacheDB[T, DB]) QueryNoCache[R any](ctx context.Context, query func(ctx context.Context, db DB) (*R, error)) (*R, error) {
	return query(ctx, cdb.db)
}

// QueryRowsNoCache 直接查询多行数据，不经过缓存（也不回填）。
func (cdb *CacheDB[T, DB]) QueryRowsNoCache[R any](ctx context.Context, query func(ctx context.Context, db DB) ([]*R, error)) ([]*R, error) {
	return query(ctx, cdb.db)
}

// Transaction 在事务中执行 fn，提交或回滚由 ant 管理。
// 事务内请使用 tx 访问数据，避免使用 cdb 自身方法（缓存层不感知事务）。
func (cdb *CacheDB[T, DB]) Transaction(ctx context.Context, fn func(tx *T) error, opts ...ant.TxOption) error {
	return cdb.db.AntContext().Tx(ctx, fn, opts...)
}

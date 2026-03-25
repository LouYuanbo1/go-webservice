package gormc

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/gormx"
)

type TypedExecFn[T any, ID comparable, PT gormx.PointerModel[T, ID]] func(
	ctx context.Context,
	s gormx.TypedSession[T, ID, PT],
) error
type TypedQueryFn[T any, ID comparable, PT gormx.PointerModel[T, ID]] func(
	ctx context.Context,
	s gormx.TypedSession[T, ID, PT],
	val PT,
) error

type TypedQueryRowsFn[T any, ID comparable, PT gormx.PointerModel[T, ID]] func(
	ctx context.Context,
	s gormx.TypedSession[T, ID, PT],
	val *[]PT,
) error

type TypedIndexQueryFn[T any, ID comparable, PT gormx.PointerModel[T, ID]] func(
	ctx context.Context,
	s gormx.TypedSession[T, ID, PT],
	val PT,
) (primaryKey ID, err error)
type TypedPrimaryQueryFn[T any, ID comparable, PT gormx.PointerModel[T, ID]] func(
	ctx context.Context,
	s gormx.TypedSession[T, ID, PT],
	val PT,
	primaryKey ID,
) error

type TypedCacheSession[T any, ID comparable, PT gormx.PointerModel[T, ID]] interface {
	GetCache(ctx context.Context, key string, val PT) error
	SetCache(ctx context.Context, key string, val PT, opts ...TTLOption) error
	DelCache(ctx context.Context, key ...string) error
	Exec(ctx context.Context, exec TypedExecFn[T, ID, PT], keys ...string) error
	Query(
		ctx context.Context,
		val PT,
		key string,
		query TypedQueryFn[T, ID, PT],
		opts ...TTLOption,
	) error
	QueryIndex(
		ctx context.Context,
		val PT,
		key string,
		keyer func(primary ID) string,
		indexQuery TypedIndexQueryFn[T, ID, PT],
		primaryQuery TypedPrimaryQueryFn[T, ID, PT],
		opts ...TTLOption,
	) error
	ExecNoCache(ctx context.Context, exec TypedExecFn[T, ID, PT]) error
	QueryNoCache(ctx context.Context, val PT, query TypedQueryFn[T, ID, PT]) error
	QueryRowsNoCache(ctx context.Context, val *[]PT, query TypedQueryRowsFn[T, ID, PT]) error
}

type typedCacheSession[T any, ID comparable, PT gormx.PointerModel[T, ID]] struct {
	db    *gormx.TypedDB
	cache cache.Client
	cfg   *Config
}

func NewTypedSessionWithCache[T any, ID comparable, PT gormx.PointerModel[T, ID]](db *gormx.TypedDB, c cache.Client, cfg *Config) TypedCacheSession[T, ID, PT] {
	return &typedCacheSession[T, ID, PT]{
		db:    db,
		cache: c,
		cfg:   cfg,
	}
}

func (tcs *typedCacheSession[T, ID, PT]) GetCache(ctx context.Context, key string, val PT) error {
	return tcs.cache.Get(ctx, key, val)
}

func (tcs *typedCacheSession[T, ID, PT]) SetCache(ctx context.Context, key string, val PT, opts ...TTLOption) error {
	return tcs.cache.Set(ctx, key, val, tcs.ttlBuilder(opts...).value)
}

func (tcs *typedCacheSession[T, ID, PT]) DelCache(ctx context.Context, key ...string) error {
	return tcs.cache.Del(ctx, key...)
}

// Cache-Aside
func (tcs *typedCacheSession[T, ID, PT]) Exec(ctx context.Context, exec TypedExecFn[T, ID, PT], keys ...string) error {
	s := gormx.GetSession[T, ID, PT](tcs.db)
	err := exec(ctx, s)
	if err != nil {
		return err
	}
	return tcs.cache.Del(ctx, keys...)
}

func (tcs *typedCacheSession[T, ID, PT]) Query(
	ctx context.Context,
	val PT,
	key string,
	query TypedQueryFn[T, ID, PT],
	opts ...TTLOption,
) error {
	return tcs.cache.Take(ctx, val, key, func(cachedVal any) error {
		s := gormx.GetSession[T, ID, PT](tcs.db)
		return query(ctx, s, cachedVal.(PT))
	}, tcs.ttlBuilder(opts...).value)
}

// 可能需要调整过期时间的关系避免缓存击穿或者雪崩
func (tcs *typedCacheSession[T, ID, PT]) QueryIndex(
	ctx context.Context,
	val PT, // 最终数据的存放指针
	key string, // 索引缓存键
	keyer func(primary ID) string, // 根据主键生成主键缓存键
	indexQuery TypedIndexQueryFn[T, ID, PT], // 通过索引查主键并填充数据
	primaryQuery TypedPrimaryQueryFn[T, ID, PT], // 通过主键查数据
	opts ...TTLOption,
) error {
	var primaryKey ID
	var foundPrimaryKeyFromDB bool // 标记是否从数据库直接获取了数据

	// 1. 尝试获取索引缓存（存储主键）
	if err := tcs.cache.Take(ctx, &primaryKey, key,
		func(cachedVal any) error { // cachedVal 实际上是 *ID
			// 从数据库通过索引获取主键，并填充完整数据到 val
			s := gormx.GetSession[T, ID, PT](tcs.db)
			pk, err := indexQuery(ctx, s, val)
			if err != nil {
				return err
			}
			primaryKey = pk
			// 将主键写入索引缓存（让 Take 自动写入）
			// cachedVal 类型为 *ID，需解引用赋值
			*(cachedVal.(*ID)) = primaryKey
			foundPrimaryKeyFromDB = true
			// 手动将完整数据写入主键缓存，过期时间略长于索引缓存
			return tcs.cache.Set(ctx, keyer(primaryKey), val,
				tcs.ttlBuilder(opts...).value+tcs.cfg.CacheSafeGapBetweenIndexAndPrimary)
		},
		tcs.ttlBuilder(opts...).value,
	); err != nil {
		return err
	}

	// 2. 如果已经从数据库获取了数据（索引缓存未命中），直接返回
	if foundPrimaryKeyFromDB {
		return nil
	}

	// 3. 索引缓存命中，得到 primaryKey，现在通过主键缓存获取完整数据
	return tcs.cache.Take(ctx, val, keyer(primaryKey), func(v any) error {
		// 主键缓存未命中时，通过数据库查询并填充
		s := gormx.GetSession[T, ID, PT](tcs.db)
		return primaryQuery(ctx, s, v.(PT), primaryKey)
	}, tcs.ttlBuilder(opts...).value)
}

func (tcs *typedCacheSession[T, ID, PT]) ExecNoCache(ctx context.Context, exec TypedExecFn[T, ID, PT]) error {
	s := gormx.GetSession[T, ID, PT](tcs.db)
	return exec(ctx, s)
}

func (tcs *typedCacheSession[T, ID, PT]) QueryNoCache(ctx context.Context, val PT, query TypedQueryFn[T, ID, PT]) error {
	s := gormx.GetSession[T, ID, PT](tcs.db)
	return query(ctx, s, val)
}

func (tcs *typedCacheSession[T, ID, PT]) QueryRowsNoCache(ctx context.Context, val *[]PT, query TypedQueryRowsFn[T, ID, PT]) error {
	s := gormx.GetSession[T, ID, PT](tcs.db)
	return query(ctx, s, val)
}
package gormc

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/gormx/gen"
)

type TypedExecFn[T any, ID comparable, PT gen.PointerModel[T, ID]] func(
	ctx context.Context,
	conn gen.Session[T, ID, PT],
) error
type TypedQueryFn[T any, ID comparable, PT gen.PointerModel[T, ID]] func(
	ctx context.Context,
	conn gen.Session[T, ID, PT],
	val PT,
) error
type TypedIndexQueryFn[T any, ID comparable, PT gen.PointerModel[T, ID]] func(
	ctx context.Context,
	conn gen.Session[T, ID, PT],
	val PT,
) (primaryKey ID, err error)
type TypedPrimaryQueryFn[T any, ID comparable, PT gen.PointerModel[T, ID]] func(
	ctx context.Context,
	conn gen.Session[T, ID, PT],
	val PT,
	primaryKey ID,
) error

type TypedCachedConn[T any, ID comparable, PT gen.PointerModel[T, ID]] interface {
	GetCache(ctx context.Context, key string, val PT) error
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
}

type typedCachedConn[T any, ID comparable, PT gen.PointerModel[T, ID]] struct {
	conn  gen.Session[T, ID, PT]
	cache cache.Client
	cfg   *Config
}

func NewTypedConnWithCache[T any, ID comparable, PT gen.PointerModel[T, ID]](db gen.Session[T, ID, PT], c cache.Client, cfg *Config) TypedCachedConn[T, ID, PT] {
	return &typedCachedConn[T, ID, PT]{
		conn:  db,
		cache: c,
		cfg:   cfg,
	}
}

func (tcc *typedCachedConn[T, ID, PT]) GetCache(ctx context.Context, key string, val PT) error {
	return tcc.cache.Get(ctx, key, val)
}

func (tcc *typedCachedConn[T, ID, PT]) DelCache(ctx context.Context, key ...string) error {
	return tcc.cache.Del(ctx, key...)
}

// Cache-Aside
func (tcc *typedCachedConn[T, ID, PT]) Exec(ctx context.Context, exec TypedExecFn[T, ID, PT], keys ...string) error {
	err := exec(ctx, tcc.conn)
	if err != nil {
		return err
	}
	return tcc.cache.Del(ctx, keys...)
}

func (tcc *typedCachedConn[T, ID, PT]) Query(
	ctx context.Context,
	val PT,
	key string,
	query TypedQueryFn[T, ID, PT],
	opts ...TTLOption,
) error {
	return tcc.cache.Take(ctx, val, key, func(val any) error {
		return query(ctx, tcc.conn, val.(PT))
	}, tcc.ttlBuilder(opts...).value)
}

// 可能需要调整过期时间的关系避免缓存击穿或者雪崩
func (tcc *typedCachedConn[T, ID, PT]) QueryIndex(
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
	if err := tcc.cache.Take(ctx, &primaryKey, key,
		func(cachedVal any) error { // cachedVal 实际上是 *ID
			// 从数据库通过索引获取主键，并填充完整数据到 val
			pk, err := indexQuery(ctx, tcc.conn, val)
			if err != nil {
				return err
			}
			primaryKey = pk

			// 将主键写入索引缓存（让 Take 自动写入）
			// cachedVal 类型为 *ID，需解引用赋值
			*(cachedVal.(*ID)) = primaryKey

			foundPrimaryKeyFromDB = true

			// 手动将完整数据写入主键缓存，过期时间略长于索引缓存
			return tcc.cache.Set(ctx, keyer(primaryKey), val,
				tcc.ttlBuilder(opts...).value+tcc.cfg.CacheSafeGapBetweenIndexAndPrimary)
		},
		tcc.ttlBuilder(opts...).value,
	); err != nil {
		return err
	}

	// 2. 如果已经从数据库获取了数据（索引缓存未命中），直接返回
	if foundPrimaryKeyFromDB {
		return nil
	}

	// 3. 索引缓存命中，得到 primaryKey，现在通过主键缓存获取完整数据
	return tcc.cache.Take(ctx, val, keyer(primaryKey), func(v any) error {
		// 主键缓存未命中时，通过数据库查询并填充
		return primaryQuery(ctx, tcc.conn, v.(PT), primaryKey)
	}, tcc.ttlBuilder(opts...).value)
}

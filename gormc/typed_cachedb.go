package gormc

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/gormx"
)

type TypedCacheDB struct {
	db    *gormx.TypedDB
	cache cache.Client
	cfg   *Config
}

func NewTypedCacheDB(db *gormx.TypedDB, c cache.Client, cfg *Config) *TypedCacheDB {
	return &TypedCacheDB{
		db:    db,
		cache: c,
		cfg:   cfg,
	}
}

func (tcdb *TypedCacheDB) GetCache(ctx context.Context, key string, val any) error {
	return tcdb.cache.Get(ctx, key, val)
}

func (tcdb *TypedCacheDB) DelCache(ctx context.Context, key ...string) error {
	return tcdb.cache.Del(ctx, key...)
}

func (tcdb *TypedCacheDB) Transaction(ctx context.Context, fn func(ctx context.Context, txDB *gormx.TypedDB) error) error {
	return tcdb.db.Transaction(ctx, fn)
}

func GetTypedCacheSession[T any, ID comparable, PT gormx.PointerModel[T, ID]](tcdb *TypedCacheDB) TypedCacheSession[T, ID, PT] {
	return NewTypedSessionWithCache[T, ID, PT](tcdb.db, tcdb.cache, tcdb.cfg)
}

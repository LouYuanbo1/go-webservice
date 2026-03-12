package gormc

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/gormx"
)

type CachedConn interface {
}

type cachedConn struct {
	conn  gormx.Conn
	cache cache.Cache
}

func NewConnWithCache(db gormx.Conn, c cache.Cache) CachedConn {
	return &cachedConn{
		conn:  db,
		cache: c,
	}
}

func (c *cachedConn) DelCache(ctx context.Context, key ...string) error {
	return c.cache.Del(ctx, key...)
}

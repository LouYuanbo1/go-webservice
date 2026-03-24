package gormx

import (
	"context"

	"gorm.io/gorm"
)

type (
// contextTxKey struct{}
)

type Conn interface {
	Session
	//Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	Transaction(ctx context.Context, fn func(ctx context.Context, s Session) error) error
}

func NewConn(db *gorm.DB) *conn {
	return &conn{Session: NewSession(db), db: db}
}

type conn struct {
	Session
	db *gorm.DB
}

func (c *conn) GetDBWithContext(ctx context.Context) *gorm.DB {
	return c.Session.GetDBWithContext(ctx)
}

func (c *conn) Create(ctx context.Context, model any, opts ...ConflictOption) error {
	return c.Session.Create(ctx, model, opts...)
}

func (c *conn) CreateInBatches(ctx context.Context, models any, batchSize int, opts ...ConflictOption) error {
	return c.Session.CreateInBatches(ctx, models, batchSize, opts...)
}

func (c *conn) GetByID(ctx context.Context, dest any, id any) error {
	return c.Session.GetByID(ctx, dest, id)
}

func (c *conn) GetByStructFilter(ctx context.Context, dest any, filter any) error {
	return c.Session.GetByStructFilter(ctx, dest, filter)
}
func (c *conn) GetByMapFilter(ctx context.Context, dest any, filter map[string]any) error {
	return c.Session.GetByMapFilter(ctx, dest, filter)
}
func (c *conn) FindByIDs(ctx context.Context, dest any, ids any, opts ...OrderOption) error {
	return c.Session.FindByIDs(ctx, dest, ids, opts...)
}
func (c *conn) FindByStructFilter(ctx context.Context, dest any, filter any, opts ...OrderOption) error {
	return c.Session.FindByStructFilter(ctx, dest, filter, opts...)
}
func (c *conn) FindByMapFilter(ctx context.Context, dest any, filter map[string]any, opts ...OrderOption) error {
	return c.Session.FindByMapFilter(ctx, dest, filter, opts...)
}

func (c *conn) FindByPage(ctx context.Context, dest any, primaryKey string, page, pageSize int, opts ...OrderOption) error {
	return c.Session.FindByPage(ctx, dest, primaryKey, page, pageSize, opts...)
}

func (c *conn) FindByCursor(ctx context.Context, dest any, primaryKey string, cursor any, limit int) error {
	return c.Session.FindByCursor(ctx, dest, primaryKey, cursor, limit)
}

func (c *conn) FindInBatches(
	ctx context.Context,
	dest any,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
	opts ...OrderOption,
) error {
	return c.Session.FindInBatches(ctx, dest, batchSize, callback, opts...)
}

func (c *conn) FindInBatchesByStructFilter(
	ctx context.Context,
	dest any,
	filter any,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
	opts ...OrderOption,
) error {
	return c.Session.FindInBatchesByStructFilter(ctx, dest, filter, batchSize, callback, opts...)
}

func (c *conn) FindInBatchesByMapFilter(
	ctx context.Context,
	dest any,
	filter map[string]any,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
	opts ...OrderOption,
) error {
	return c.Session.FindInBatchesByMapFilter(ctx, dest, filter, batchSize, callback, opts...)
}

func (c *conn) Update(ctx context.Context, updateData any) error {
	return c.Session.Update(ctx, updateData)
}

func (c *conn) UpdatesByStructFilter(ctx context.Context, filter any, updateData any) error {
	return c.Session.UpdatesByStructFilter(ctx, filter, updateData)
}
func (c *conn) UpdatesByMapFilter(ctx context.Context, model any, filter map[string]any, updateData map[string]any) error {
	return c.Session.UpdatesByMapFilter(ctx, model, filter, updateData)
}
func (c *conn) DeleteByID(ctx context.Context, model any, id any) error {
	return c.Session.DeleteByID(ctx, model, id)
}
func (c *conn) DeleteByIDs(ctx context.Context, model any, ids any) error {
	return c.Session.DeleteByIDs(ctx, model, ids)
}
func (c *conn) DeleteByStructFilter(ctx context.Context, model any, filter any) error {
	return c.Session.DeleteByStructFilter(ctx, model, filter)
}
func (c *conn) DeleteByMapFilter(ctx context.Context, model any, filter map[string]any) error {
	return c.Session.DeleteByMapFilter(ctx, model, filter)
}

func (c *conn) Transaction(ctx context.Context, fn func(ctx context.Context, s Session) error) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		s := NewSession(tx)
		return fn(ctx, s)
	})
}

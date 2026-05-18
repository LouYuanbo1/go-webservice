package gormx

import (
	"context"

	"gorm.io/gorm"
)

type DB interface {
	Session
	Transaction(ctx context.Context, fn func(ctx context.Context, s Session) error) error
}

func NewDB(db *gorm.DB) *xdb {
	return &xdb{Session: NewSession(db), db: db}
}

type xdb struct {
	Session
	db *gorm.DB
}

func (xdb *xdb) GetDBWithContext(ctx context.Context) *gorm.DB {
	return xdb.Session.GetDBWithContext(ctx)
}

func (xdb *xdb) Create(ctx context.Context, model any, opts ...ConflictOption) error {
	return xdb.Session.Create(ctx, model, opts...)
}

func (xdb *xdb) CreateInBatches(ctx context.Context, models any, batchSize int, opts ...ConflictOption) error {
	return xdb.Session.CreateInBatches(ctx, models, batchSize, opts...)
}

func (xdb *xdb) GetByID(ctx context.Context, dest any, id any) error {
	return xdb.Session.GetByID(ctx, dest, id)
}

func (xdb *xdb) GetByStructFilter(ctx context.Context, dest any, filter any) error {
	return xdb.Session.GetByStructFilter(ctx, dest, filter)
}

func (xdb *xdb) FindByIDs(ctx context.Context, dest any, ids any, opts ...OrderOption) error {
	return xdb.Session.FindByIDs(ctx, dest, ids, opts...)
}
func (xdb *xdb) FindByStructFilter(ctx context.Context, dest any, filter any, opts ...OrderOption) error {
	return xdb.Session.FindByStructFilter(ctx, dest, filter, opts...)
}

func (xdb *xdb) FindByPage(ctx context.Context, dest any, page, pageSize int, opts ...OrderOption) error {
	return xdb.Session.FindByPage(ctx, dest, page, pageSize, opts...)
}

func (xdb *xdb) FindByCursor(ctx context.Context, dest any, cursor any, limit int) error {
	return xdb.Session.FindByCursor(ctx, dest, cursor, limit)
}

func (xdb *xdb) FindInBatches(
	ctx context.Context,
	dest any,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
	opts ...OrderOption,
) error {
	return xdb.Session.FindInBatches(ctx, dest, batchSize, callback, opts...)
}

func (xdb *xdb) FindInBatchesByStructFilter(
	ctx context.Context,
	dest any,
	filter any,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
	opts ...OrderOption,
) error {
	return xdb.Session.FindInBatchesByStructFilter(ctx, dest, filter, batchSize, callback, opts...)
}

func (xdb *xdb) Update(ctx context.Context, updateData any) error {
	return xdb.Session.Update(ctx, updateData)
}

func (xdb *xdb) UpdatesByStructFilter(ctx context.Context, filter any, updateData any) error {
	return xdb.Session.UpdatesByStructFilter(ctx, filter, updateData)
}

func (xdb *xdb) DeleteByID(ctx context.Context, model any, id any) error {
	return xdb.Session.DeleteByID(ctx, model, id)
}
func (xdb *xdb) DeleteByIDs(ctx context.Context, model any, ids any) error {
	return xdb.Session.DeleteByIDs(ctx, model, ids)
}
func (xdb *xdb) DeleteByStructFilter(ctx context.Context, model any, filter any) error {
	return xdb.Session.DeleteByStructFilter(ctx, model, filter)
}

func (xdb *xdb) Transaction(ctx context.Context, fn func(ctx context.Context, s Session) error) error {
	return xdb.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		s := NewSession(tx)
		return fn(ctx, s)
	})
}

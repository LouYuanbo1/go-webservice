package gormx

import (
	"context"

	"gorm.io/gorm"
)

type GormX[T any, ID comparable, PT PointerModel[T, ID]] interface {
	GetDBWithContext(ctx context.Context) *gorm.DB
	InTransaction(ctx context.Context) bool
	Create(ctx context.Context, model PT, opts ...ConflictOption) error
	CreateInBatches(ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) error
	GetByID(ctx context.Context, id ID) (PT, error)
	GetByStructFilter(ctx context.Context, filter PT) (PT, error)
	GetByMapFilter(ctx context.Context, filter map[string]any) (PT, error)
	FindByIDs(ctx context.Context, ids []ID, opts ...OrderOption) ([]PT, error)
	FindByStructFilter(ctx context.Context, filter PT, opts ...OrderOption) ([]PT, error)
	FindByMapFilter(ctx context.Context, filter map[string]any, opts ...OrderOption) ([]PT, error)
	FindByPage(ctx context.Context, page, pageSize int, opts ...OrderOption) ([]PT, error)
	FindByCursor(ctx context.Context, cursor ID, limit int) ([]PT, ID, bool, error)
	FindInBatches(ctx context.Context, batchSize int, callback func(ctx context.Context, batch int, ptrModels []PT) error, opts ...OrderOption) error
	FindInBatchesByStructFilter(ctx context.Context, filter PT, batchSize int, callback func(ctx context.Context, batch int, ptrModels []PT) error, opts ...OrderOption) error
	FindInBatchesByMapFilter(ctx context.Context, filter map[string]any, batchSize int, callback func(ctx context.Context, batch int, ptrModels []PT) error, opts ...OrderOption) error
	Update(ctx context.Context, updateData PT) error
	UpdateByStructFilter(ctx context.Context, filter PT, updateData PT) error
	UpdateByMapFilter(ctx context.Context, filter map[string]any, updateData map[string]any) error
	DeleteByID(ctx context.Context, id ID) error
	DeleteByIDs(ctx context.Context, ids []ID) error
	DeleteByStructFilter(ctx context.Context, filter PT) error
	DeleteByMapFilter(ctx context.Context, filter map[string]any) error
}

type gormX[T any, ID comparable, PT PointerModel[T, ID]] struct {
	db *gorm.DB
}

func NewGormX[T any, ID comparable, PT PointerModel[T, ID]](db *gorm.DB) *gormX[T, ID, PT] {
	return &gormX[T, ID, PT]{db: db}
}

func (gx *gormX[T, ID, PT]) GetDBWithContext(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if !ok {
		return gx.db.WithContext(ctx)
	}
	return tx.WithContext(ctx)
}
func (gx *gormX[T, ID, PT]) InTransaction(ctx context.Context) bool {
	_, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	return ok
}

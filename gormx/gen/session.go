package gen

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/gormx"
	"gorm.io/gorm"
)

type Session[T any, ID comparable, PT PointerModel[T, ID]] interface {
	GetDBWithContext(ctx context.Context) *gorm.DB
	InTransaction(ctx context.Context) bool
	Create(ctx context.Context, model PT, opts ...gormx.ConflictOption) error
	CreateInBatches(ctx context.Context, models []PT, batchSize int, opts ...gormx.ConflictOption) error
	GetByID(ctx context.Context, id ID) (PT, error)
	GetByStructFilter(ctx context.Context, filter PT) (PT, error)
	GetByMapFilter(ctx context.Context, filter map[string]any) (PT, error)
	FindByIDs(ctx context.Context, ids []ID, opts ...gormx.OrderOption) ([]PT, error)
	FindByStructFilter(ctx context.Context, filter PT, opts ...gormx.OrderOption) ([]PT, error)
	FindByMapFilter(ctx context.Context, filter map[string]any, opts ...gormx.OrderOption) ([]PT, error)
	FindByPage(ctx context.Context, page, pageSize int, opts ...gormx.OrderOption) ([]PT, error)
	FindByCursor(ctx context.Context, cursor ID, limit int) ([]PT, ID, bool, error)
	FindInBatches(
		ctx context.Context,
		batchSize int,
		callback func(ctx context.Context, batch int, models []PT) error, opts ...gormx.OrderOption) error
	FindInBatchesByStructFilter(
		ctx context.Context,
		filter PT,
		batchSize int,
		callback func(ctx context.Context, batch int, models []PT) error, opts ...gormx.OrderOption) error
	FindInBatchesByMapFilter(
		ctx context.Context,
		filter map[string]any,
		batchSize int,
		callback func(ctx context.Context, batch int, models []PT) error, opts ...gormx.OrderOption) error
	Update(ctx context.Context, updateData PT) error
	UpdateByStructFilter(ctx context.Context, filter PT, updateData PT) error
	UpdateByMapFilter(ctx context.Context, filter map[string]any, updateData map[string]any) error
	DeleteByID(ctx context.Context, id ID) error
	DeleteByIDs(ctx context.Context, ids []ID) error
	DeleteByStructFilter(ctx context.Context, filter PT) error
	DeleteByMapFilter(ctx context.Context, filter map[string]any) error
}

type genSession[T any, ID comparable, PT PointerModel[T, ID]] struct {
	gormx.Session
	db *gorm.DB
}

func NewSession[T any, ID comparable, PT PointerModel[T, ID]](db *gorm.DB) Session[T, ID, PT] {
	return &genSession[T, ID, PT]{db: db}
}

func (s *genSession[T, ID, PT]) GetDBWithContext(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if !ok {
		return s.db.WithContext(ctx)
	}
	return tx.WithContext(ctx)
}
func (s *genSession[T, ID, PT]) InTransaction(ctx context.Context) bool {
	_, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	return ok
}

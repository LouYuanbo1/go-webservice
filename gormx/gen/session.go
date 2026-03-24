package gen

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/gormx"
	"gorm.io/gorm"
)

type Session[T any, ID comparable, PT PointerModel[T, ID]] interface {
	GetDBWithContext(ctx context.Context) *gorm.DB
	Create(ctx context.Context, model PT, opts ...gormx.ConflictOption) error
	CreateInBatches(ctx context.Context, models []PT, batchSize int, opts ...gormx.ConflictOption) error
	GetByID(ctx context.Context, dest PT, id ID) error
	GetByStructFilter(ctx context.Context, dest PT, filter PT) error
	GetByMapFilter(ctx context.Context, dest PT, filter map[string]any) error
	FindByIDs(ctx context.Context, dest *[]PT, ids []ID, opts ...gormx.OrderOption) error
	FindByStructFilter(ctx context.Context, dest *[]PT, filter PT, opts ...gormx.OrderOption) error
	FindByMapFilter(ctx context.Context, dest *[]PT, filter map[string]any, opts ...gormx.OrderOption) error
	FindByPage(ctx context.Context, dest *[]PT, page, pageSize int, opts ...gormx.OrderOption) error
	FindByCursor(ctx context.Context, dest *[]PT, cursor ID, limit int) (ID, bool, error)
	FindInBatches(
		ctx context.Context,
		batchSize int,
		callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error, opts ...gormx.OrderOption) error
	FindInBatchesByStructFilter(
		ctx context.Context,
		filter PT,
		batchSize int,
		callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error, opts ...gormx.OrderOption) error
	FindInBatchesByMapFilter(
		ctx context.Context,
		filter map[string]any,
		batchSize int,
		callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error, opts ...gormx.OrderOption) error
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
}

func NewSession[T any, ID comparable, PT PointerModel[T, ID]](db *gorm.DB) Session[T, ID, PT] {
	return &genSession[T, ID, PT]{Session: gormx.NewSession(db)}
}

func (g *genSession[T, ID, PT]) GetDBWithContext(ctx context.Context) *gorm.DB {
	return g.Session.GetDBWithContext(ctx)
}

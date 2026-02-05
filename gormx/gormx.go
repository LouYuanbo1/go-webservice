package gormx

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/gormx/internal"
	"github.com/LouYuanbo1/go-webservice/gormx/model"
	"gorm.io/gorm"
)

type GormX[T any, ID comparable, PT model.PointerModel[T, ID]] interface {
	DB() *gorm.DB
	InTransaction(ctx context.Context) bool
	Create(ctx context.Context, model PT) error
	CreateInBatches(ctx context.Context, models []PT, batchSize int) error
	FirstOrCreate(ctx context.Context, model PT) (PT, error)
	GetByID(ctx context.Context, id ID) (PT, error)
	FindByIDs(ctx context.Context, ids []ID) ([]PT, error)
	GetByStructFilter(ctx context.Context, filter PT) (PT, error)
	FindByStructFilter(ctx context.Context, filter PT) ([]PT, error)
	GetByMapFilter(ctx context.Context, filter map[string]any) (PT, error)
	FindByMapFilter(ctx context.Context, filter map[string]any) ([]PT, error)
	FindByPage(ctx context.Context, page, pageSize int) ([]PT, error)
	FindByCursor(ctx context.Context, cursor ID, pageSize int) ([]PT, ID, bool, error)
	Update(ctx context.Context, updateData PT) error
	UpdateByStructFilter(ctx context.Context, filter PT, updateData PT) error
	UpdateByMapFilter(ctx context.Context, filter map[string]any, updateData map[string]any) error
	DeleteByID(ctx context.Context, id ID) error
	DeleteByIDs(ctx context.Context, ids []ID) error
	DeleteByStructFilter(ctx context.Context, filter PT) error
	DeleteByMapFilter(ctx context.Context, filter map[string]any) error
}

func NewGormX[T any, ID comparable, PT model.PointerModel[T, ID]](db *gorm.DB) GormX[T, ID, PT] {
	return internal.NewGormX[T, ID, PT](db)
}

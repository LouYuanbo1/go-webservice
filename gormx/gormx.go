package gormx

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/gormx/internal"
	"github.com/LouYuanbo1/go-webservice/gormx/model"
	"gorm.io/gorm"
)

type GormX[T any, PT model.PointerModel[T]] interface {
	Create(ctx context.Context, ptrModel PT) error
	CreateInBatches(ctx context.Context, ptrModels []PT, batchSize int) error
	GetByID(ctx context.Context, id uint64) (PT, error)
	GetByIDs(ctx context.Context, ids []uint64) ([]PT, error)
	GetByStructFields(ctx context.Context, structModel PT) ([]PT, error)
	GetByMapFields(ctx context.Context, mapFields map[string]any) ([]PT, error)
	GetByPage(ctx context.Context, page, pageSize uint64) ([]PT, error)
	GetByCursor(ctx context.Context, cursor, pageSize uint64) ([]PT, uint64, bool, error)
	Update(ctx context.Context, ptrModel PT) error
	DeleteByID(ctx context.Context, id uint64) error
	DeleteByIDs(ctx context.Context, ids []uint64) error
}

func NewGormx[T any, PT model.PointerModel[T]](db *gorm.DB) GormX[T, PT] {
	return internal.NewGormx[T, PT](db)
}

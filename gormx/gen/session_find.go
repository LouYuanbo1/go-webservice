package gen

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/LouYuanbo1/go-webservice/gormx"
	"gorm.io/gorm"
)

func (g *genSession[T, ID, PT]) FindByIDs(ctx context.Context, ids []ID, opts ...gormx.OrderOption) ([]PT, error) {
	if len(ids) == 0 {
		log.Printf("find by ids failed : %s", gormx.WarnEmptyIDSlice)
		return nil, nil
	}
	for _, id := range ids {
		if IsZero(id) {
			log.Printf("find by ids failed, index: %v : %s", id, gormx.WarnInvalidID)
			return nil, nil
		}
	}

	ptrs := make([]PT, 0, len(ids))
	if err := g.Session.FindByIDs(ctx, &ptrs, ids, opts...); err != nil {
		return nil, err
	}
	return ptrs, nil
}

func (g *genSession[T, ID, PT]) FindByStructFilter(ctx context.Context, filter PT, opts ...gormx.OrderOption) ([]PT, error) {
	ptrs := make([]PT, 0, 50)
	if err := g.Session.FindByStructFilter(ctx, &ptrs, filter, opts...); err != nil {
		return nil, err
	}
	return ptrs, nil
}

func (g *genSession[T, ID, PT]) FindByMapFilter(ctx context.Context, filter map[string]any, opts ...gormx.OrderOption) ([]PT, error) {
	ptrs := make([]PT, 0, 50)
	if err := g.Session.FindByMapFilter(ctx, &ptrs, filter, opts...); err != nil {
		return nil, err
	}
	return ptrs, nil
}

func (g *genSession[T, ID, PT]) FindByPage(ctx context.Context, page, pageSize int, opts ...gormx.OrderOption) ([]PT, error) {
	var model T
	ptr := PT(&model)
	primaryKey := ptr.PrimaryKey()
	ptrs := make([]PT, 0, pageSize)
	if err := g.Session.FindByPage(ctx, &ptrs, primaryKey, page, pageSize, opts...); err != nil {
		return nil, err
	}
	return ptrs, nil
}

func (g *genSession[T, ID, PT]) FindByCursor(ctx context.Context, cursor ID, limit int) ([]PT, ID, bool, error) {
	if limit <= 0 {
		log.Printf("find by cursor failed : %s", gormx.WarnInvalidLimit)
		return nil, cursor, false, nil
	}
	if IsZero(cursor) {
		log.Printf("find by cursor failed : %s", gormx.WarnInvalidID)
		return nil, cursor, false, nil
	}

	var model T
	ptr := PT(&model)
	primaryKey := ptr.PrimaryKey()
	ptrs := make([]PT, 0, limit+1)

	err := g.Session.FindByCursor(ctx, &ptrs, primaryKey, cursor, limit)
	if err != nil {
		return nil, cursor, false, err
	}
	hasMore := len(ptrs) > limit
	if hasMore {
		ptrs = ptrs[:limit]
	}
	newCursor := cursor
	if len(ptrs) > 0 {
		newCursor = ptrs[len(ptrs)-1].GetID()
	}
	return ptrs, newCursor, hasMore, nil
}

// 并非零成本的调用，断言类型为 []PT
func (g *genSession[T, ID, PT]) FindInBatches(
	ctx context.Context,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error,
	opts ...gormx.OrderOption,
) error {
	ptrs := make([]PT, 0, batchSize)
	err := g.Session.FindInBatches(ctx, &ptrs, batchSize,
		func(ctx context.Context, tx *gorm.DB, batch int, models any) error {
			typedModels, ok := models.(*[]PT)
			if !ok {
				return errorx.NewWithDetails(
					gormx.ErrInvalidTypeAssertion,
					"gormx",
					"FindInBatchesByStructFilter",
					fmt.Sprintf("unexpected type: %T", models),
					nil,
				)
			}
			return callback(ctx, tx, batch, *typedModels)
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// 并非零成本的调用，断言类型为 []PT
func (g *genSession[T, ID, PT]) FindInBatchesByStructFilter(
	ctx context.Context,
	filter PT,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, ptrModels []PT) error,
	opts ...gormx.OrderOption,
) error {
	ptrs := make([]PT, 0, batchSize)
	err := g.Session.FindInBatchesByStructFilter(ctx, &ptrs, filter, batchSize,
		func(ctx context.Context, tx *gorm.DB, batch int, models any) error {
			typedModels, ok := models.(*[]PT)
			if !ok {
				return errorx.NewWithDetails(
					gormx.ErrInvalidTypeAssertion,
					"gormx",
					"FindInBatchesByStructFilter",
					fmt.Sprintf("unexpected type: %T", models),
					nil,
				)
			}
			return callback(ctx, tx, batch, *typedModels)
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// 并非零成本的调用，断言类型为 []PT
func (g *genSession[T, ID, PT]) FindInBatchesByMapFilter(
	ctx context.Context,
	filter map[string]any,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, ptrModels []PT) error,
	opts ...gormx.OrderOption,
) error {
	ptrs := make([]PT, 0, batchSize)
	err := g.Session.FindInBatchesByMapFilter(ctx, &ptrs, filter, batchSize,
		func(ctx context.Context, tx *gorm.DB, batch int, models any) error {
			typedModels, ok := models.(*[]PT)
			if !ok {
				return errorx.NewWithDetails(
					gormx.ErrInvalidTypeAssertion,
					"gormx",
					"FindInBatchesByMapFilter",
					fmt.Sprintf("unexpected type: %T", models),
					nil,
				)
			}
			return callback(ctx, tx, batch, *typedModels)
		},
	)
	if err != nil {
		return err
	}
	return nil
}

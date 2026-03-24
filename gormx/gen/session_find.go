package gen

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/LouYuanbo1/go-webservice/gormx"
	"gorm.io/gorm"
)

func (g *genSession[T, ID, PT]) FindByIDs(ctx context.Context, dest *[]PT, ids []ID, opts ...gormx.OrderOption) error {
	if len(ids) == 0 {
		log.Printf("find by ids failed : %s", gormx.WarnEmptyIDSlice)
		return nil
	}
	for _, id := range ids {
		if IsZero(id) {
			log.Printf("find by ids failed, index: %v : %s", id, gormx.WarnInvalidID)
			return nil
		}
	}
	return g.Session.FindByIDs(ctx, dest, ids, opts...)
}

func (g *genSession[T, ID, PT]) FindByStructFilter(ctx context.Context, dest *[]PT, filter PT, opts ...gormx.OrderOption) error {
	return g.Session.FindByStructFilter(ctx, dest, filter, opts...)
}

func (g *genSession[T, ID, PT]) FindByMapFilter(ctx context.Context, dest *[]PT, filter map[string]any, opts ...gormx.OrderOption) error {
	return g.Session.FindByMapFilter(ctx, dest, filter, opts...)
}

func (g *genSession[T, ID, PT]) FindByPage(ctx context.Context, dest *[]PT, page, pageSize int, opts ...gormx.OrderOption) error {
	var model T
	ptr := PT(&model)
	primaryKey := ptr.PrimaryKey()
	return g.Session.FindByPage(ctx, dest, primaryKey, page, pageSize, opts...)
}

func (g *genSession[T, ID, PT]) FindByCursor(ctx context.Context, dest *[]PT, cursor ID, limit int) (ID, bool, error) {
	if limit <= 0 {
		log.Printf("find by cursor failed : %s", gormx.WarnInvalidLimit)
		return cursor, false, nil
	}
	if IsZero(cursor) {
		log.Printf("find by cursor failed : %s", gormx.WarnInvalidID)
		return cursor, false, nil
	}

	var model T
	ptr := PT(&model)
	primaryKey := ptr.PrimaryKey()

	err := g.Session.FindByCursor(ctx, dest, primaryKey, cursor, limit)
	if err != nil {
		return cursor, false, err
	}
	hasMore := len(*dest) > limit
	if hasMore {
		*dest = (*dest)[:limit]
	}
	newCursor := cursor
	if len(*dest) > 0 {
		newCursor = (*dest)[len(*dest)-1].GetID()
	}
	return newCursor, hasMore, nil
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
	callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error,
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
	callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error,
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

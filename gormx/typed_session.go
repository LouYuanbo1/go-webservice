package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm"
)

type TypedSession[T any, ID comparable, PT PointerModel[T, ID]] interface {
	GetDBWithContext(ctx context.Context) *gorm.DB
	Create(ctx context.Context, model PT, opts ...ConflictOption) error
	CreateInBatches(ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) error
	GetByID(ctx context.Context, dest PT, id ID) error
	GetByStructFilter(ctx context.Context, dest PT, filter PT) error
	GetByMapFilter(ctx context.Context, dest PT, filter map[string]any) error
	FindByIDs(ctx context.Context, dest *[]PT, ids []ID, opts ...OrderOption) error
	FindByStructFilter(ctx context.Context, dest *[]PT, filter PT, opts ...OrderOption) error
	FindByMapFilter(ctx context.Context, dest *[]PT, filter map[string]any, opts ...OrderOption) error
	FindByPage(ctx context.Context, dest *[]PT, page, pageSize int, opts ...OrderOption) error
	FindByCursor(ctx context.Context, dest *[]PT, cursor ID, limit int) (newCursor ID, hasMore bool, err error)
	FindInBatches(
		ctx context.Context,
		batchSize int,
		callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error, opts ...OrderOption) error
	FindInBatchesByStructFilter(
		ctx context.Context,
		filter PT,
		batchSize int,
		callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error, opts ...OrderOption) error
	FindInBatchesByMapFilter(
		ctx context.Context,
		filter map[string]any,
		batchSize int,
		callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error, opts ...OrderOption) error
	Update(ctx context.Context, updateData PT) error
	UpdateByStructFilter(ctx context.Context, filter PT, updateData PT) error
	UpdateByMapFilter(ctx context.Context, filter map[string]any, updateData map[string]any) error
	DeleteByID(ctx context.Context, id ID) error
	DeleteByIDs(ctx context.Context, ids ...ID) error
	DeleteByStructFilter(ctx context.Context, filter PT) error
	DeleteByMapFilter(ctx context.Context, filter map[string]any) error
}

type typedSession[T any, ID comparable, PT PointerModel[T, ID]] struct {
	Session
}

func NewTypedSession[T any, ID comparable, PT PointerModel[T, ID]](db *gorm.DB) TypedSession[T, ID, PT] {
	return &typedSession[T, ID, PT]{Session: NewSession(db)}
}

func (ts *typedSession[T, ID, PT]) GetDBWithContext(ctx context.Context) *gorm.DB {
	return ts.Session.GetDBWithContext(ctx)
}

func (ts *typedSession[T, ID, PT]) Create(ctx context.Context, model PT, opts ...ConflictOption) error {
	return ts.Session.Create(ctx, model, opts...)
}

func (ts *typedSession[T, ID, PT]) CreateInBatches(ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) error {
	if len(models) == 0 {
		// 空切片属于合法操作（0 行插入），静默成功更符合批量操作语义
		log.Printf("skipped create in batches: %s", WarnEmptyModelsSlice)
		return nil
	}
	return ts.Session.CreateInBatches(ctx, models, batchSize, opts...)
}

func (ts *typedSession[T, ID, PT]) GetByID(ctx context.Context, dest PT, id ID) error {
	if IsZero(id) {
		log.Printf("get by id failed : %s", WarnInvalidID)
		return nil
	}
	return ts.Session.GetByID(ctx, dest, id)
}

func (ts *typedSession[T, ID, PT]) GetByStructFilter(ctx context.Context, dest PT, filter PT) error {
	return ts.Session.GetByStructFilter(ctx, dest, filter)
}

func (ts *typedSession[T, ID, PT]) GetByMapFilter(ctx context.Context, dest PT, filter map[string]any) error {
	return ts.Session.GetByMapFilter(ctx, dest, filter)
}

func (ts *typedSession[T, ID, PT]) FindByIDs(ctx context.Context, dest *[]PT, ids []ID, opts ...OrderOption) error {
	if len(ids) == 0 {
		log.Printf("find by ids failed : %s", WarnEmptyIDSlice)
		return nil
	}
	for _, id := range ids {
		if IsZero(id) {
			log.Printf("find by ids failed, index: %v : %s", id, WarnInvalidID)
			return nil
		}
	}
	return ts.Session.FindByIDs(ctx, dest, ids, opts...)
}

func (ts *typedSession[T, ID, PT]) FindByStructFilter(ctx context.Context, dest *[]PT, filter PT, opts ...OrderOption) error {
	return ts.Session.FindByStructFilter(ctx, dest, filter, opts...)
}

func (ts *typedSession[T, ID, PT]) FindByMapFilter(ctx context.Context, dest *[]PT, filter map[string]any, opts ...OrderOption) error {
	return ts.Session.FindByMapFilter(ctx, dest, filter, opts...)
}

func (ts *typedSession[T, ID, PT]) FindByPage(ctx context.Context, dest *[]PT, page, pageSize int, opts ...OrderOption) error {
	model := PT(new(T))
	return ts.Session.FindByPage(ctx, dest, model.PrimaryKey(), page, pageSize, opts...)
}

func (ts *typedSession[T, ID, PT]) FindByCursor(ctx context.Context, dest *[]PT, cursor ID, limit int) (newCursor ID, hasMore bool, err error) {
	if limit <= 0 {
		log.Printf("find by cursor failed : %s", WarnInvalidLimit)
		return cursor, false, nil
	}

	model := PT(new(T))

	err = ts.Session.FindByCursor(ctx, dest, model.PrimaryKey(), cursor, limit+1)
	if err != nil {
		return cursor, false, err
	}
	hasMore = len(*dest) > limit
	if hasMore {
		*dest = (*dest)[:limit]
	}
	newCursor = cursor
	if len(*dest) > 0 {
		newCursor = (*dest)[len(*dest)-1].GetID()
	}
	return newCursor, hasMore, nil
}

// 并非零成本的调用，断言类型为 []PT
func (ts *typedSession[T, ID, PT]) FindInBatches(
	ctx context.Context,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error,
	opts ...OrderOption,
) error {
	ptrs := make([]PT, 0, batchSize)
	err := ts.Session.FindInBatches(ctx, &ptrs, batchSize,
		func(ctx context.Context, tx *gorm.DB, batch int, models any) error {
			typedModels, ok := models.(*[]PT)
			if !ok {
				return errorx.NewWithDetails(
					ErrInvalidTypeAssertion,
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
func (ts *typedSession[T, ID, PT]) FindInBatchesByStructFilter(
	ctx context.Context,
	filter PT,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error,
	opts ...OrderOption,
) error {
	ptrs := make([]PT, 0, batchSize)
	err := ts.Session.FindInBatchesByStructFilter(ctx, &ptrs, filter, batchSize,
		func(ctx context.Context, tx *gorm.DB, batch int, models any) error {
			typedModels, ok := models.(*[]PT)
			if !ok {
				return errorx.NewWithDetails(
					ErrInvalidTypeAssertion,
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
func (ts *typedSession[T, ID, PT]) FindInBatchesByMapFilter(
	ctx context.Context,
	filter map[string]any,
	batchSize int,
	callback func(ctx context.Context, tx *gorm.DB, batch int, models []PT) error,
	opts ...OrderOption,
) error {
	ptrs := make([]PT, 0, batchSize)
	err := ts.Session.FindInBatchesByMapFilter(ctx, &ptrs, filter, batchSize,
		func(ctx context.Context, tx *gorm.DB, batch int, models any) error {
			typedModels, ok := models.(*[]PT)
			if !ok {
				return errorx.NewWithDetails(
					ErrInvalidTypeAssertion,
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

func (ts *typedSession[T, ID, PT]) Update(ctx context.Context, updateData PT) error {
	return ts.Session.Update(ctx, updateData)
}

func (ts *typedSession[T, ID, PT]) UpdateByStructFilter(ctx context.Context, filter PT, updateData PT) error {
	return ts.Session.UpdatesByStructFilter(ctx, filter, updateData)
}

func (ts *typedSession[T, ID, PT]) UpdateByMapFilter(ctx context.Context, filter map[string]any, updateData map[string]any) error {
	model := PT(new(T))
	return ts.Session.UpdatesByMapFilter(ctx, model, filter, updateData)
}

func (ts *typedSession[T, ID, PT]) DeleteByID(ctx context.Context, id ID) error {
	if IsZero(id) {
		log.Printf("delete by id failed : %s", WarnInvalidID)
		return nil
	}

	model := PT(new(T))
	return ts.Session.DeleteByID(ctx, model, id)
}

func (ts *typedSession[T, ID, PT]) DeleteByIDs(ctx context.Context, ids ...ID) error {
	if len(ids) == 0 {
		log.Printf("delete by ids failed : %s", WarnEmptyIDSlice)
		return nil
	}
	for _, id := range ids {
		if IsZero(id) {
			log.Printf("delete by ids failed, index: %v : %s", id, WarnInvalidID)
			return nil
		}
	}

	model := PT(new(T))
	return ts.Session.DeleteByIDs(ctx, model, ids)
}

func (ts *typedSession[T, ID, PT]) DeleteByStructFilter(ctx context.Context, filter PT) error {
	model := PT(new(T))
	return ts.Session.DeleteByStructFilter(ctx, model, filter)
}

func (ts *typedSession[T, ID, PT]) DeleteByMapFilter(ctx context.Context, filter map[string]any) error {
	model := PT(new(T))
	return ts.Session.DeleteByMapFilter(ctx, model, filter)
}

package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/gormx/model"
	"gorm.io/gorm"
)

type gormX[T any, PT model.PointerModel[T]] struct {
	db *gorm.DB
}

func NewGormx[T any, PT model.PointerModel[T]](db *gorm.DB) *gormX[T, PT] {
	return &gormX[T, PT]{db: db}
}

func (gx *gormX[T, PT]) getDBWithContext(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if !ok {
		return gx.db.WithContext(ctx)
	}
	return tx.WithContext(ctx)
}

func (gx *gormX[T, PT]) Create(ctx context.Context, ptrModel PT) error {
	if ptrModel == nil {
		return fmt.Errorf("create %s failed, ptrModel is nil", ptrModel.TableName())
	}

	fmt.Printf("type ptrModel: %T", ptrModel)

	result := gx.getDBWithContext(ctx).
		Create(ptrModel)
	if result.Error != nil {
		log.Printf("create %s failed, error: %v", ptrModel.TableName(), result.Error)
		return fmt.Errorf("create %s failed, error: %v", ptrModel.TableName(), result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("create %s failed, no rows affected", ptrModel.TableName())
		return fmt.Errorf("create %s failed, no rows affected", ptrModel.TableName())
	}
	return nil
}

func (gx *gormX[T, PT]) CreateInBatches(ctx context.Context, ptrModels []PT, batchSize int) error {
	// 检查ptrModels是否为空
	if len(ptrModels) == 0 || batchSize <= 0 {
		var model T
		ptr := PT(&model)
		log.Printf("create %s in batchs failed, no models provided", ptr.TableName())
		return nil
	}
	result := gx.getDBWithContext(ctx).
		CreateInBatches(ptrModels, batchSize)
	if result.Error != nil {
		log.Printf("create %s in batchs failed, error: %v", ptrModels[0].TableName(), result.Error)
		return fmt.Errorf("create %s in batchs failed, error: %v", ptrModels[0].TableName(), result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("create %s in batchs failed, no rows affected", ptrModels[0].TableName())
		return fmt.Errorf("create %s in batchs failed, no rows affected", ptrModels[0].TableName())
	}
	return nil
}

func (gx *gormX[T, PT]) GetByID(ctx context.Context, id uint64) (PT, error) {
	var model T
	ptrModel := PT(&model)

	if id == 0 {
		log.Printf("get %s by id %d failed, id must be greater than 0", ptrModel.TableName(), id)
		return nil, nil
	}

	result := gx.getDBWithContext(ctx).
		Where(fmt.Sprintf("%s = ?", ptrModel.GetPrimaryKey()), id).
		First(ptrModel)
	if result.Error != nil {
		log.Printf("get %s by id %d failed, error: %v", ptrModel.TableName(), id, result.Error)
		return nil, fmt.Errorf("get %s by id %d failed, error: %v", ptrModel.TableName(), id, result.Error)
	}
	return ptrModel, nil
}

func (gx *gormX[T, PT]) GetByIDs(ctx context.Context, ids []uint64) ([]PT, error) {
	var model T
	ptrModel := PT(&model)

	if len(ids) == 0 {
		log.Printf("get %s by ids %v failed, no ids provided", ptrModel.TableName(), ids)
		return nil, nil
	}

	fmt.Printf("type model: %T, type ptrModel: %T", model, ptrModel)

	ptrModels := make([]PT, 0, len(ids))

	fmt.Printf("type ptrModels: %T", ptrModels)

	result := gx.getDBWithContext(ctx).
		Where(fmt.Sprintf("%s IN ?", ptrModel.GetPrimaryKey()), ids).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("get %s by ids %v failed, error: %v", ptrModel.TableName(), ids, result.Error)
		return nil, fmt.Errorf("get %s by ids %v failed, error: %v", ptrModel.TableName(), ids, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("get %s by ids %v failed, no rows affected", ptrModel.TableName(), ids)
		return nil, fmt.Errorf("get %s by ids %v failed, no rows affected", ptrModel.TableName(), ids)
	}
	return ptrModels, nil
}

func (gx *gormX[T, PT]) GetByStructFields(ctx context.Context, structModel PT) ([]PT, error) {
	if structModel == nil {
		log.Printf("get %s by structModel %v failed, structModel is nil", structModel.TableName(), structModel)
		return nil, nil
	}
	ptrModels := make([]PT, 0, 10)
	result := gx.getDBWithContext(ctx).
		Where(structModel).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("get %s by structModel %v failed, error: %v", structModel.TableName(), structModel, result.Error)
		return nil, fmt.Errorf("get %s by structModel %v failed, error: %v", structModel.TableName(), structModel, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("get %s by structModel %v failed, no rows affected", structModel.TableName(), structModel)
		return nil, fmt.Errorf("get %s by structModel %v failed, no rows affected", structModel.TableName(), structModel)
	}
	return ptrModels, nil
}

func (gx *gormX[T, PT]) GetByMapFields(ctx context.Context, mapFields map[string]any) ([]PT, error) {
	var model T
	ptrModel := PT(&model)
	if mapFields == nil {
		log.Printf("get %s by mapFields %v failed, mapFields is nil", ptrModel.TableName(), mapFields)
		return nil, nil
	}
	ptrModels := make([]PT, 0, 10)
	result := gx.getDBWithContext(ctx).
		Where(mapFields).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("get %s by mapFields %v failed, error: %v", ptrModel.TableName(), mapFields, result.Error)
		return nil, fmt.Errorf("get %s by mapFields %v failed, error: %v", ptrModel.TableName(), mapFields, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("get %s by mapFields %v failed, no rows affected", ptrModel.TableName(), mapFields)
		return nil, fmt.Errorf("get %s by mapFields %v failed, no rows affected", ptrModel.TableName(), mapFields)
	}
	return ptrModels, nil
}

func (gx *gormX[T, PT]) GetByPage(ctx context.Context, page, pageSize uint64) ([]PT, error) {
	var model T
	ptrModel := PT(&model)

	if page <= 0 || pageSize <= 0 {
		log.Printf("get %s by page %d, pageSize %d failed, page and pageSize must be greater than 0", ptrModel.TableName(), page, pageSize)
		return nil, nil
	}

	ptrModels := make([]PT, 0, pageSize)

	result := gx.getDBWithContext(ctx).
		Order(fmt.Sprintf("%s ASC", ptrModel.GetPrimaryKey())).
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("get %s by page %d, pageSize %d failed, error: %v", ptrModel.TableName(), page, pageSize, result.Error)
		return nil, fmt.Errorf("get %s by page %d, pageSize %d failed, error: %v", ptrModel.TableName(), page, pageSize, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("get %s by page %d, pageSize %d failed, no rows affected", ptrModel.TableName(), page, pageSize)
		return nil, fmt.Errorf("get %s by page %d, pageSize %d failed, no rows affected", ptrModel.TableName(), page, pageSize)
	}
	return ptrModels, nil
}

func (gx *gormX[T, PT]) GetByCursor(ctx context.Context, cursor, pageSize uint64) ([]PT, uint64, bool, error) {
	var model T
	ptrModel := PT(&model)

	if pageSize <= 0 {
		log.Printf("get %s by cursor %d, pageSize %d failed, pageSize must be greater than 0", ptrModel.TableName(), cursor, pageSize)
		return nil, cursor, false, nil
	}
	limit := pageSize + 1

	ptrModels := make([]PT, 0, pageSize)
	result := gx.getDBWithContext(ctx).
		Where(fmt.Sprintf("%s = 0 OR %s > ?", ptrModel.GetPrimaryKey(), ptrModel.GetPrimaryKey()), cursor).
		Order(fmt.Sprintf("%s ASC", ptrModel.GetPrimaryKey())).
		Limit(int(limit)).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("get %s by cursor %d, pageSize %d failed, error: %v", ptrModel.TableName(), cursor, pageSize, result.Error)
		return nil, cursor, false, fmt.Errorf("get %s by cursor %d, pageSize %d failed, error: %v", ptrModel.TableName(), cursor, pageSize, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("get %s by cursor %d, pageSize %d failed, no rows affected", ptrModel.TableName(), cursor, pageSize)
		return nil, cursor, false, fmt.Errorf("get %s by cursor %d, pageSize %d failed, no rows affected", ptrModel.TableName(), cursor, pageSize)
	}
	hasMore := uint64(len(ptrModels)) > pageSize
	if hasMore {
		ptrModels = ptrModels[:pageSize]
	}
	newCursor := cursor
	if len(ptrModels) > 0 {
		newCursor = ptrModels[len(ptrModels)-1].GetID()
	}
	return ptrModels, newCursor, hasMore, nil
}

func (gx *gormX[T, PT]) Update(ctx context.Context, ptrModel PT) error {
	if ptrModel == nil {
		var model T
		ptr := PT(&model)
		log.Printf("update %s failed, ptrModel is nil", ptr.TableName())
		return nil
	}
	fmt.Printf("type ptrModel: %T", ptrModel)

	result := gx.getDBWithContext(ctx).
		Updates(ptrModel)
	if result.Error != nil {
		log.Printf("update %s failed, error: %v", ptrModel.TableName(), result.Error)
		return fmt.Errorf("update %s failed, error: %v", ptrModel.TableName(), result.Error)
	}
	return nil
}

func (gx *gormX[T, PT]) DeleteByID(ctx context.Context, id uint64) error {
	var model T
	pt := PT(&model)

	if id == 0 {
		log.Printf("delete %s failed, id is 0", pt.TableName())
		return nil
	}
	result := gx.getDBWithContext(ctx).
		Where(fmt.Sprintf("%s = ?", pt.GetPrimaryKey()), id).
		Delete(pt)
	if result.Error != nil {
		log.Printf("delete %s failed, error: %v", pt.TableName(), result.Error)
		return fmt.Errorf("delete %s failed, error: %v", pt.TableName(), result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete %s failed, no rows affected", pt.TableName())
		return fmt.Errorf("delete %s failed, no rows affected", pt.TableName())
	}
	return nil
}

func (gx *gormX[T, PT]) DeleteByIDs(ctx context.Context, ids []uint64) error {
	var model T
	pt := PT(&model)

	if len(ids) == 0 {
		log.Printf("delete %s by ids failed, no ids provided", pt.TableName())
		return nil
	}
	result := gx.getDBWithContext(ctx).
		Where(fmt.Sprintf("%s IN ?", pt.GetPrimaryKey()), ids).
		Delete(pt)
	if result.Error != nil {
		log.Printf("delete %s by ids %v failed, error: %v", pt.TableName(), ids, result.Error)
		return fmt.Errorf("delete %s by ids %v failed, error: %v", pt.TableName(), ids, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete %s by ids %v failed, no rows affected", pt.TableName(), ids)
		return fmt.Errorf("delete %s by ids %v failed, no rows affected", pt.TableName(), ids)
	}
	return nil
}

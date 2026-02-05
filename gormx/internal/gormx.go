package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/gormx/model"
	"gorm.io/gorm"
)

type gormX[T any, ID comparable, PT model.PointerModel[T, ID]] struct {
	db *gorm.DB
}

func NewGormX[T any, ID comparable, PT model.PointerModel[T, ID]](db *gorm.DB) *gormX[T, ID, PT] {
	return &gormX[T, ID, PT]{db: db}
}

func (gx *gormX[T, ID, PT]) getDBWithContext(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if !ok {
		return gx.db.WithContext(ctx)
	}
	return tx.WithContext(ctx)
}

func (gx *gormX[T, ID, PT]) DB() *gorm.DB {
	return gx.db
}

func (gx *gormX[T, ID, PT]) InTransaction(ctx context.Context) bool {
	_, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	return ok
}

func (gx *gormX[T, ID, PT]) Create(ctx context.Context, model PT) error {
	if model == nil {
		return fmt.Errorf("create failed, model is nil")
	}

	tableName := model.TableName()

	result := gx.getDBWithContext(ctx).
		Create(model)
	if result.Error != nil {
		log.Printf("create failed. table: %s, error: %v", tableName, result.Error)
		return fmt.Errorf("create failed. table: %s, error: %v", tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("create failed. table: %s, no rows affected", tableName)
		//return fmt.Errorf("create failed. table: %s, no rows affected", tableName)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) CreateInBatches(ctx context.Context, models []PT, batchSize int) error {
	// 参数校验
	if batchSize <= 0 {
		return fmt.Errorf("create in batches failed: batchSize must be positive, got %d", batchSize)
	}
	if len(models) == 0 {
		// 空切片属于合法操作（0 行插入），静默成功更符合批量操作语义
		log.Printf("skipped batch create: empty models slice")
		return nil
	}

	tableName := models[0].TableName()

	result := gx.getDBWithContext(ctx).
		CreateInBatches(models, batchSize)
	if result.Error != nil {
		log.Printf("create in batches failed. table: %s, error: %v", tableName, result.Error)
		return fmt.Errorf("create in batches failed. table: %s, error: %v", tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("create in batches failed. table: %s, no rows affected", tableName)
		//return fmt.Errorf("create in batches failed. table: %s, no rows affected", tableName)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) FirstOrCreate(ctx context.Context, model PT) (PT, error) {
	if model == nil {
		log.Printf("first or create failed, model is nil")
		return nil, nil
	}

	tableName := model.TableName()

	result := gx.getDBWithContext(ctx).
		FirstOrCreate(model)
	if result.Error != nil {
		log.Printf("first or create failed. table: %s, error: %v", tableName, result.Error)
		return nil, fmt.Errorf("first or create failed. table: %s, error: %v", tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("first or create failed. table: %s, no rows affected", tableName)
		//return nil, fmt.Errorf("first or create failed. table: %s, no rows affected", tableName)
	}
	return model, nil
}

func (gx *gormX[T, ID, PT]) GetByID(ctx context.Context, id ID) (PT, error) {

	if model.IsZero(id) {
		log.Printf("get by id failed, id not be zero value")
		return nil, nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.getDBWithContext(ctx).
		First(ptr, id)
	if result.Error != nil {
		log.Printf("get by id failed. table: %s, error: %v", tableName, result.Error)
		return nil, fmt.Errorf("get by id failed. table: %s, error: %v", tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by id failed. table: %s, no rows affected", tableName)
		//return nil, fmt.Errorf("get by id failed. table: %s, no rows affected", tableName)
	}
	return ptr, nil
}

func (gx *gormX[T, ID, PT]) FindByIDs(ctx context.Context, ids []ID) ([]PT, error) {
	if len(ids) == 0 {
		log.Printf("find by ids failed, no ids provided")
		return nil, nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()
	ptrModels := make([]PT, 0, len(ids))

	result := gx.getDBWithContext(ctx).
		Find(&ptr, ids)
	if result.Error != nil {
		log.Printf("find by ids failed. table: %s, error: %v", tableName, result.Error)
		return nil, fmt.Errorf("find by ids failed. table: %s, error: %v", tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by ids failed. table: %s, no rows affected", tableName)
		//return nil, fmt.Errorf("find by ids failed. table: %s, no rows affected", tableName)
	}
	return ptrModels, nil
}

func (gx *gormX[T, ID, PT]) GetByStructFilter(ctx context.Context, filter PT) (PT, error) {
	if filter == nil {
		log.Printf("get by struct filter failed, filter is nil")
		return nil, nil
	}

	var model T
	ptrModel := PT(&model)
	tableName := ptrModel.TableName()

	result := gx.getDBWithContext(ctx).
		Where(filter).
		First(ptrModel)
	if result.Error != nil {
		log.Printf("get by struct filter failed. table: %s, error: %v", tableName, result.Error)
		return nil, fmt.Errorf("get by struct filter failed. table: %s, error: %v", tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by struct filter failed. table: %s, no rows affected", tableName)
		//return nil, fmt.Errorf("get %s by struct filter %v failed, no rows affected", tableName, filter)
	}
	return ptrModel, nil
}

func (gx *gormX[T, ID, PT]) FindByStructFilter(ctx context.Context, filter PT) ([]PT, error) {
	if filter == nil {
		log.Printf("find by struct filter failed, filter is nil")
		return nil, nil
	}

	ptrModels := make([]PT, 0, 50)
	tableName := filter.TableName()

	result := gx.getDBWithContext(ctx).
		Where(filter).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("find by struct filter failed. table: %s, error: %v", tableName, result.Error)
		return nil, fmt.Errorf("find by struct filter failed. table: %s, error: %v", tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by struct filter failed. table: %s, no rows affected", tableName)
		//return nil, fmt.Errorf("find %s by struct filter %v failed, no rows affected", tableName, filter)
	}
	return ptrModels, nil
}

func (gx *gormX[T, ID, PT]) GetByMapFilter(ctx context.Context, filter map[string]any) (PT, error) {
	if filter == nil {
		log.Printf("get by map filter failed, filter is nil")
		return nil, nil
	}

	var model T
	ptrModel := PT(&model)
	tableName := ptrModel.TableName()

	result := gx.getDBWithContext(ctx).
		Where(filter).
		First(ptrModel)
	if result.Error != nil {
		log.Printf("get by map filter failed. table: %s, error: %v", tableName, result.Error)
		return nil, fmt.Errorf("get by map filter failed. table: %s, error: %v", tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by map filter failed. table: %s, no rows affected", tableName)
		//return nil, fmt.Errorf("get by map filter failed. table: %s, no rows affected", tableName)
	}
	return ptrModel, nil
}

func (gx *gormX[T, ID, PT]) FindByMapFilter(ctx context.Context, filter map[string]any) ([]PT, error) {
	if filter == nil {
		log.Printf("find by map filter failed, filter is nil")
		return nil, nil
	}
	if len(filter) == 0 {
		log.Printf("find by map filter failed, filter is empty")
		return nil, nil
	}

	var model T
	ptrModel := PT(&model)
	tableName := ptrModel.TableName()
	ptrModels := make([]PT, 0, 50)

	result := gx.getDBWithContext(ctx).
		Where(filter).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("find by map filter failed. table: %s, error: %v", tableName, result.Error)
		return nil, fmt.Errorf("find by map filter failed. table: %s, error: %v", tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by map filter failed. table: %s, no rows affected", tableName)
		//return nil, fmt.Errorf("find by map filter failed. table: %s, no rows affected", tableName)
	}
	return ptrModels, nil
}

func (gx *gormX[T, ID, PT]) FindByPage(ctx context.Context, page, pageSize int) ([]PT, error) {
	if page <= 0 || pageSize <= 0 {
		log.Printf("find by page %d, pageSize %d failed, page and pageSize must be greater than zero", page, pageSize)
		return nil, nil
	}

	var model T
	ptrModel := PT(&model)
	primaryKey := ptrModel.PrimaryKey()
	tableName := ptrModel.TableName()
	ptrModels := make([]PT, 0, pageSize)

	result := gx.getDBWithContext(ctx).
		Order(fmt.Sprintf("%s ASC", primaryKey)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("find by page %d, pageSize %d failed. table: %s, error: %v", page, pageSize, tableName, result.Error)
		return nil, fmt.Errorf("find by page %d, pageSize %d failed. table: %s, error: %v", page, pageSize, tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by page %d, pageSize %d failed. table: %s, no rows affected", page, pageSize, tableName)
		//return nil, fmt.Errorf("find %s by page %d, pageSize %d failed, no rows affected", tableName, page, pageSize)
	}
	return ptrModels, nil
}

func (gx *gormX[T, ID, PT]) FindByCursor(ctx context.Context, cursor ID, pageSize int) ([]PT, ID, bool, error) {
	if pageSize <= 0 {
		log.Printf("find by cursor failed, pageSize %d must be greater than zero", pageSize)
		return nil, cursor, false, nil
	}

	if model.IsZero(cursor) {
		log.Printf("find by cursor failed, cursor %v must be not zero value", cursor)
		return nil, cursor, false, nil
	}

	limit := pageSize + 1
	var model T
	ptrModel := PT(&model)
	primaryKey := ptrModel.PrimaryKey()
	tableName := ptrModel.TableName()
	ptrModels := make([]PT, 0, limit)

	result := gx.getDBWithContext(ctx).
		Where(fmt.Sprintf("%s = 0 OR %s > ?", primaryKey, primaryKey), cursor).
		Order(fmt.Sprintf("%s ASC", primaryKey)).
		Limit(limit).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("find by cursor %v, pageSize %d failed. table: %s, error: %v", cursor, pageSize, tableName, result.Error)
		return nil, cursor, false, fmt.Errorf("find by cursor %v, pageSize %d failed. table: %s, error: %v", cursor, pageSize, tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by cursor %v, pageSize %d failed. table: %s, no rows affected", cursor, pageSize, tableName)
		//return nil, cursor, false, fmt.Errorf("find by cursor %v, pageSize %d failed. table: %s, no rows affected", cursor, pageSize, tableName)
	}
	hasMore := len(ptrModels) > pageSize
	if hasMore {
		ptrModels = ptrModels[:pageSize]
	}
	newCursor := cursor
	if len(ptrModels) > 0 {
		newCursor = ptrModels[len(ptrModels)-1].ID()
	}
	return ptrModels, newCursor, hasMore, nil
}

func (gx *gormX[T, ID, PT]) Update(ctx context.Context, updateData PT) error {
	if updateData == nil {
		log.Printf("update failed, update data must be not nil")
		return nil
	}

	result := gx.getDBWithContext(ctx).
		Updates(updateData)
	if result.Error != nil {
		log.Printf("update failed. table: %s, error: %v", updateData.TableName(), result.Error)
		return fmt.Errorf("update failed. table: %s, error: %v", updateData.TableName(), result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("update failed. table: %s, no rows affected", updateData.TableName())
		//return fmt.Errorf("update failed. table: %s, no rows affected", updateData.TableName())
	}
	return nil
}

func (gx *gormX[T, ID, PT]) UpdateByStructFilter(ctx context.Context, filter PT, updateData PT) error {
	if updateData == nil {
		log.Printf("update by struct filter failed, update data must be not nil")
		return nil
	}
	if filter == nil {
		log.Printf("update by struct filter failed, filter must be not nil")
		return nil
	}

	result := gx.getDBWithContext(ctx).
		Where(filter).
		Updates(updateData)
	if result.Error != nil {
		log.Printf("update by struct filter %v failed. table: %s error: %v", filter, updateData.TableName(), result.Error)
		return fmt.Errorf("update by struct filter %v failed. table: %s error: %v", filter, updateData.TableName(), result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("update by struct filter %v failed. table: %s, no rows affected", filter, updateData.TableName())
		//return fmt.Errorf("update by struct filter %v failed. table: %s, no rows affected", filter, updateData.TableName())
	}
	return nil
}

func (gx *gormX[T, ID, PT]) UpdateByMapFilter(ctx context.Context, filter map[string]any, updateData map[string]any) error {
	if updateData == nil {
		log.Printf("update by map filter failed, update data must be not nil")
		return nil
	}
	if len(updateData) == 0 {
		log.Printf("update by map filter failed, update data must be not empty")
		return nil
	}
	if filter == nil {
		log.Printf("update by map filter failed, filter must be not nil")
		return nil
	}
	if len(filter) == 0 {
		log.Printf("update by map filter failed, filter must be not empty")
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.getDBWithContext(ctx).
		Where(filter).
		Updates(updateData)
	if result.Error != nil {
		log.Printf("update by map filter %v failed. table: %s error: %v", filter, tableName, result.Error)
		return fmt.Errorf("update by map filter %v failed. table: %s error: %v", filter, tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("update by map filter %v failed. table: %s, no rows affected", filter, tableName)
		//return fmt.Errorf("update by map filter %v failed. table: %s, no rows affected", filter, tableName)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) DeleteByID(ctx context.Context, id ID) error {
	if model.IsZero(id) {
		log.Printf("delete by id %v failed, id must be not zero value", id)
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.getDBWithContext(ctx).
		Delete(ptr, id)
	if result.Error != nil {
		log.Printf("delete by id %v failed. table: %s error: %v", id, tableName, result.Error)
		return fmt.Errorf("delete by id %v failed. table: %s error: %v", id, tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by id %v failed. table: %s no rows affected", id, tableName)
		//return fmt.Errorf("delete by id %v failed. table: %s no rows affected", id, tableName)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) DeleteByIDs(ctx context.Context, ids []ID) error {
	if len(ids) == 0 {
		log.Printf("delete by ids failed, no ids provided")
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.getDBWithContext(ctx).
		Delete(ptr, ids)
	if result.Error != nil {
		log.Printf("delete by ids %v failed. table: %s error: %v", ids, tableName, result.Error)
		return fmt.Errorf("delete by ids %v failed. table: %s error: %v", ids, tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by ids %v failed. table: %s no rows affected", ids, tableName)
		//return fmt.Errorf("delete by ids %v failed. table: %s no rows affected", ids, tableName)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) DeleteByStructFilter(ctx context.Context, filter PT) error {
	if filter == nil {
		log.Printf("delete by struct filter failed, filter must be not nil")
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.getDBWithContext(ctx).
		Where(filter).
		Delete(ptr)
	if result.Error != nil {
		log.Printf("delete by struct filter %v failed. table: %s error: %v", filter, tableName, result.Error)
		return fmt.Errorf("delete by struct filter %v failed. table: %s error: %v", filter, tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by struct filter %v failed. table: %s no rows affected", filter, tableName)
		//return fmt.Errorf("delete by struct filter %v failed. table: %s no rows affected", filter, tableName)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) DeleteByMapFilter(ctx context.Context, filter map[string]any) error {
	if filter == nil {
		log.Printf("delete by map filter failed, filter must be not nil")
		return nil
	}
	if len(filter) == 0 {
		log.Printf("delete by map filter failed, filter must be not empty")
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.getDBWithContext(ctx).
		Where(filter).
		Delete(ptr)
	if result.Error != nil {
		log.Printf("delete by map filter %v failed. table: %s error: %v", filter, tableName, result.Error)
		return fmt.Errorf("delete by map filter %v failed. table: %s error: %v", filter, tableName, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by map filter %v failed. table: %s no rows affected", filter, tableName)
		//return fmt.Errorf("delete by map filter %v failed. table: %s no rows affected", filter, tableName)
	}
	return nil
}

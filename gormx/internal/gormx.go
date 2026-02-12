package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/gormx/errors"
	"github.com/LouYuanbo1/go-webservice/gormx/model"
	"github.com/LouYuanbo1/go-webservice/gormx/options"
	"gorm.io/gorm"
)

type gormX[T any, ID comparable, PT model.PointerModel[T, ID]] struct {
	db *gorm.DB
}

func NewGormX[T any, ID comparable, PT model.PointerModel[T, ID]](db *gorm.DB) *gormX[T, ID, PT] {
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

func (gx *gormX[T, ID, PT]) Create(ctx context.Context, model PT, opts ...options.ConflictOption) error {
	if model == nil {
		log.Printf("create failed : %s", errors.WarnInvalidModel)
		return nil
	}

	tableName := model.TableName()
	var result *gorm.DB
	// 应用冲突选项
	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			Create(model)
		if result.Error != nil {
			return errors.New(
				errors.ErrCreateFailed,
				"Create",
				tableName,
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("create failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
		}
		return nil
	}

	clauseConflict, err := gx.clauseOnConflictBuilder(opts...)
	if err != nil {
		return errors.New(
			errors.ErrInvalidOnConflictClause,
			"Create",
			tableName,
			err,
		)
	}

	result = gx.GetDBWithContext(ctx).
		Clauses(clauseConflict).
		Create(model)
	if result.Error != nil {
		return errors.New(
			errors.ErrCreateFailed,
			"Create(Upsert)",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("create(upsert) failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) CreateInBatches(ctx context.Context, models []PT, batchSize int, opts ...options.ConflictOption) error {
	// 参数校验
	if batchSize <= 0 {
		log.Printf("create in batches failed : %s", errors.WarnInvalidBatchSize)
		return nil
	}
	if len(models) == 0 {
		// 空切片属于合法操作（0 行插入），静默成功更符合批量操作语义
		log.Printf("skipped create in batches: %s", errors.WarnEmptyModelsSlice)
		return nil
	}

	tableName := models[0].TableName()
	var result *gorm.DB

	if len(opts) == 0 {
		result = gx.GetDBWithContext(ctx).
			CreateInBatches(models, batchSize)
		if result.Error != nil {
			log.Printf("create in batches failed. table: %s, error: %v", tableName, result.Error)
			return errors.New(
				errors.ErrCreateFailed,
				"CreateInBatches",
				tableName,
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("create in batches failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
		}
		return nil
	}

	// 应用冲突选项
	clauseConflict, err := gx.clauseOnConflictBuilder(opts...)
	if err != nil {
		return errors.New(
			errors.ErrInvalidOnConflictClause,
			"CreateInBatches",
			tableName,
			err,
		)
	}

	result = gx.GetDBWithContext(ctx).
		Clauses(clauseConflict).
		CreateInBatches(models, batchSize)
	if result.Error != nil {
		log.Printf("create(upsert) in batches failed. table: %s, error: %v", tableName, result.Error)
		return errors.New(
			errors.ErrCreateFailed,
			"CreateInBatches(Upsert)",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("create in batches failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) GetByID(ctx context.Context, id ID) (PT, error) {

	if model.IsZero(id) {
		log.Printf("get by id failed : %s", errors.WarnInvalidID)
		return nil, nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.GetDBWithContext(ctx).
		First(ptr, id)
	if result.Error != nil {
		log.Printf("get by id failed. table: %s, error: %v", tableName, result.Error)
		return nil, errors.New(
			errors.ErrQueryFailed,
			"GetByID",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by id failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return ptr, nil
}

func (gx *gormX[T, ID, PT]) FindByIDs(ctx context.Context, ids []ID, opts ...options.OrderOption) ([]PT, error) {
	if len(ids) == 0 {
		log.Printf("find by ids failed : %s", errors.WarnEmptyIDsSlice)
		return nil, nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()
	ptrModels := make([]PT, 0, len(ids))
	var result *gorm.DB

	if len(opts) == 0 {

		result = gx.GetDBWithContext(ctx).
			Find(&ptrModels, ids)
		if result.Error != nil {
			log.Printf("find by ids failed. table: %s, error: %v", tableName, result.Error)
			return nil, errors.New(
				errors.ErrQueryFailed,
				"FindByIDs",
				tableName,
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by ids failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
		}

		return ptrModels, nil
	}

	clauseOrder := gx.clauseOrderBuilder(opts...)

	result = gx.GetDBWithContext(ctx).
		Order(clauseOrder).
		Find(&ptrModels, ids)
	if result.Error != nil {
		log.Printf("find by ids failed. table: %s, error: %v", tableName, result.Error)
		return nil, errors.New(
			errors.ErrQueryFailed,
			"FindByIDs(Order)",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by ids (order) failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}

	return ptrModels, nil
}

func (gx *gormX[T, ID, PT]) GetByStructFilter(ctx context.Context, filter PT) (PT, error) {
	if filter == nil {
		log.Printf("get by struct filter failed : %s", errors.WarnInvalidFilter)
		return nil, nil
	}

	var model T
	ptrModel := PT(&model)
	tableName := ptrModel.TableName()

	result := gx.GetDBWithContext(ctx).
		Where(filter).
		First(ptrModel)
	if result.Error != nil {
		log.Printf("get by struct filter failed. table: %s, error: %v", tableName, result.Error)
		return nil, errors.New(
			errors.ErrQueryFailed,
			"GetByStructFilter",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by struct filter failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return ptrModel, nil
}

func (gx *gormX[T, ID, PT]) FindByStructFilter(ctx context.Context, filter PT, opts ...options.OrderOption) ([]PT, error) {
	if filter == nil {
		log.Printf("find by struct filter failed : %s", errors.WarnInvalidFilter)
		return nil, nil
	}

	ptrModels := make([]PT, 0, 50)
	tableName := filter.TableName()
	var result *gorm.DB

	if len(opts) == 0 {

		result = gx.GetDBWithContext(ctx).
			Where(filter).
			Find(&ptrModels)
		if result.Error != nil {
			log.Printf("find by struct filter failed. table: %s, error: %v", tableName, result.Error)
			return nil, errors.New(
				errors.ErrQueryFailed,
				"FindByStructFilter",
				tableName,
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by struct filter failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
		}

		return ptrModels, nil
	}

	clauseOrder := gx.clauseOrderBuilder(opts...)

	result = gx.GetDBWithContext(ctx).
		Where(filter).
		Order(clauseOrder).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("find by struct filter (order) failed. table: %s, error: %v", tableName, result.Error)
		return nil, errors.New(
			errors.ErrQueryFailed,
			"FindByStructFilter(Order)",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by ids (order) failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}

	return ptrModels, nil
}

func (gx *gormX[T, ID, PT]) GetByMapFilter(ctx context.Context, filter map[string]any) (PT, error) {
	if filter == nil {
		log.Printf("get by map filter failed : %s", errors.WarnInvalidFilter)
		return nil, nil
	}
	if len(filter) == 0 {
		log.Printf("get by map filter failed : %s", errors.WarnInvalidFilter)
		return nil, nil
	}

	var model T
	ptrModel := PT(&model)
	tableName := ptrModel.TableName()

	result := gx.GetDBWithContext(ctx).
		Where(filter).
		First(ptrModel)
	if result.Error != nil {
		log.Printf("get by map filter failed. table: %s, error: %v", tableName, result.Error)
		return nil, errors.New(
			errors.ErrQueryFailed,
			"GetByMapFilter",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by map filter failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return ptrModel, nil
}

func (gx *gormX[T, ID, PT]) FindByMapFilter(ctx context.Context, filter map[string]any, opts ...options.OrderOption) ([]PT, error) {
	if filter == nil {
		log.Printf("find by map filter failed : %s", errors.WarnInvalidFilter)
		return nil, nil
	}
	if len(filter) == 0 {
		log.Printf("find by map filter failed : %s", errors.WarnInvalidFilter)
		return nil, nil
	}

	var model T
	ptrModel := PT(&model)
	tableName := ptrModel.TableName()
	ptrModels := make([]PT, 0, 50)
	var result *gorm.DB

	if len(opts) == 0 {

		result = gx.GetDBWithContext(ctx).
			Where(filter).
			Find(&ptrModels)
		if result.Error != nil {
			log.Printf("find by map filter failed. table: %s, error: %v", tableName, result.Error)
			return nil, errors.New(
				errors.ErrQueryFailed,
				"FindByMapFilter",
				tableName,
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by map filter failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
		}

		return ptrModels, nil
	}

	clauseOrder := gx.clauseOrderBuilder(opts...)

	result = gx.GetDBWithContext(ctx).
		Where(filter).
		Order(clauseOrder).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("find by map filter failed. table: %s, error: %v", tableName, result.Error)
		return nil, errors.New(
			errors.ErrQueryFailed,
			"FindByMapFilter(Order)",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by map filter (order) failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}

	return ptrModels, nil
}

func (gx *gormX[T, ID, PT]) FindByPage(ctx context.Context, page, pageSize int, opts ...options.OrderOption) ([]PT, error) {
	if page <= 0 || pageSize <= 0 {
		log.Printf("find by page %d, pageSize %d failed : %s", page, pageSize, errors.WarnInvalidPageParams)
		return nil, nil
	}

	var model T
	ptrModel := PT(&model)
	primaryKey := ptrModel.PrimaryKey()
	tableName := ptrModel.TableName()
	ptrModels := make([]PT, 0, pageSize)
	var result *gorm.DB

	if len(opts) == 0 {

		result = gx.GetDBWithContext(ctx).
			Order(fmt.Sprintf("%s ASC", primaryKey)).
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&ptrModels)
		if result.Error != nil {
			log.Printf("find by page %d, pageSize %d failed. table: %s, error: %v", page, pageSize, tableName, result.Error)
			return nil, errors.New(
				errors.ErrQueryFailed,
				"FindByPage",
				tableName,
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("find by page %d, pageSize %d failed. table: %s, %s", page, pageSize, tableName, errors.WarnNoRowsAffected)
		}

		return ptrModels, nil
	}

	clauseOrder := gx.clauseOrderBuilder(opts...)

	result = gx.GetDBWithContext(ctx).
		Order(clauseOrder).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("find by page %d, pageSize %d (order) failed. table: %s, error: %v", page, pageSize, tableName, result.Error)
		return nil, errors.New(
			errors.ErrQueryFailed,
			"FindByPage",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by page %d, pageSize %d (order) failed. table: %s, %s", page, pageSize, tableName, errors.WarnNoRowsAffected)
	}

	return ptrModels, nil
}

func (gx *gormX[T, ID, PT]) FindByCursor(ctx context.Context, cursor ID, limit int) ([]PT, ID, bool, error) {
	if limit <= 0 {
		log.Printf("find by cursor failed : %s", errors.WarnInvalidLimit)
		return nil, cursor, false, nil
	}

	if model.IsZero(cursor) {
		log.Printf("find by cursor failed : %s", errors.WarnInvalidID)
		return nil, cursor, false, nil
	}

	var model T
	ptrModel := PT(&model)
	primaryKey := ptrModel.PrimaryKey()
	tableName := ptrModel.TableName()
	ptrModels := make([]PT, 0, limit)

	result := gx.GetDBWithContext(ctx).
		Where(fmt.Sprintf("%s > ?", primaryKey), cursor).
		Order(fmt.Sprintf("%s ASC", primaryKey)).
		Limit(limit + 1).
		Find(&ptrModels)
	if result.Error != nil {
		log.Printf("find by cursor %v, limit %d failed. table: %s, error: %v", cursor, limit, tableName, result.Error)
		return nil, cursor, false, errors.New(
			errors.ErrQueryFailed,
			"FindByCursor",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("find by cursor %v, limit %d failed. table: %s, %s", cursor, limit, tableName, errors.WarnNoRowsAffected)
	}
	hasMore := len(ptrModels) > limit
	if hasMore {
		ptrModels = ptrModels[:limit]
	}
	newCursor := cursor
	if len(ptrModels) > 0 {
		newCursor = ptrModels[len(ptrModels)-1].GetID()
	}

	return ptrModels, newCursor, hasMore, nil
}

func (gx *gormX[T, ID, PT]) Update(ctx context.Context, updateData PT) error {
	if updateData == nil {
		log.Printf("update failed : %s", errors.WarnInvalidUpdateData)
		return nil
	}

	tableName := updateData.TableName()

	result := gx.GetDBWithContext(ctx).
		Updates(updateData)
	if result.Error != nil {
		log.Printf("update failed. table: %s, error: %v", tableName, result.Error)
		return errors.New(
			errors.ErrUpdateFailed,
			"Update",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("update failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) UpdateByStructFilter(ctx context.Context, filter PT, updateData PT) error {
	if updateData == nil {
		log.Printf("update by struct filter failed : %s", errors.WarnInvalidUpdateData)
		return nil
	}
	if filter == nil {
		log.Printf("update by struct filter failed : %s", errors.WarnInvalidFilter)
		return nil
	}

	tableName := updateData.TableName()

	result := gx.GetDBWithContext(ctx).
		Where(filter).
		Updates(updateData)
	if result.Error != nil {
		log.Printf("update by struct filter %v failed. table: %s error: %v", filter, tableName, result.Error)
		return errors.New(
			errors.ErrUpdateFailed,
			"UpdateByStructFilter",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("update by struct filter failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) UpdateByMapFilter(ctx context.Context, filter map[string]any, updateData map[string]any) error {
	if updateData == nil {
		log.Printf("update by map filter failed : %s", errors.WarnInvalidUpdateData)
		return nil
	}
	if len(updateData) == 0 {
		log.Printf("update by map filter failed : %s", errors.WarnInvalidUpdateData)
		return nil
	}
	if filter == nil {
		log.Printf("update by map filter failed : %s", errors.WarnInvalidFilter)
		return nil
	}
	if len(filter) == 0 {
		log.Printf("update by map filter failed : %s", errors.WarnInvalidFilter)
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.GetDBWithContext(ctx).
		Where(filter).
		Updates(updateData)
	if result.Error != nil {
		log.Printf("update by map filter %v failed. table: %s error: %v", filter, tableName, result.Error)
		return errors.New(
			errors.ErrUpdateFailed,
			"UpdateByMapFilter",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("update by map filter failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) DeleteByID(ctx context.Context, id ID) error {
	if model.IsZero(id) {
		log.Printf("delete by id failed : %s", errors.WarnInvalidID)
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.GetDBWithContext(ctx).
		Delete(ptr, id)
	if result.Error != nil {
		log.Printf("delete by id %v failed. table: %s, error: %v", id, tableName, result.Error)
		return errors.New(
			errors.ErrDeleteFailed,
			"DeleteByID",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by id failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) DeleteByIDs(ctx context.Context, ids []ID) error {
	if len(ids) == 0 {
		log.Printf("delete by ids failed : %s", errors.WarnEmptyIDsSlice)
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.GetDBWithContext(ctx).
		Delete(ptr, ids)
	if result.Error != nil {
		log.Printf("delete by ids %v failed. table: %s error: %v", ids, tableName, result.Error)
		return errors.New(
			errors.ErrDeleteFailed,
			"DeleteByIDs",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by ids failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) DeleteByStructFilter(ctx context.Context, filter PT) error {
	if filter == nil {
		log.Printf("delete by struct filter failed : %s", errors.WarnInvalidFilter)
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.GetDBWithContext(ctx).
		Where(filter).
		Delete(ptr)
	if result.Error != nil {
		log.Printf("delete by struct filter %v failed. table: %s error: %v", filter, tableName, result.Error)
		return errors.New(
			errors.ErrDeleteFailed,
			"DeleteByStructFilter",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by struct filter failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) DeleteByMapFilter(ctx context.Context, filter map[string]any) error {
	if filter == nil {
		log.Printf("delete by map filter failed : %s", errors.WarnInvalidFilter)
		return nil
	}
	if len(filter) == 0 {
		log.Printf("delete by map filter failed : %s", errors.WarnInvalidFilter)
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.GetDBWithContext(ctx).
		Where(filter).
		Delete(ptr)
	if result.Error != nil {
		log.Printf("delete by map filter %v failed. table: %s, error: %v", filter, tableName, result.Error)
		return errors.New(
			errors.ErrDeleteFailed,
			"DeleteByMapFilter",
			tableName,
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by map filter failed. table: %s, %s", tableName, errors.WarnNoRowsAffected)
	}
	return nil
}

package internal

import (
	"context"
	"log"

	"github.com/LouYuanbo1/go-webservice/gormx/errors"
	"github.com/LouYuanbo1/go-webservice/gormx/model"
)

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

func (gx *gormX[T, ID, PT]) GetByStructFilter(ctx context.Context, filter PT) (PT, error) {
	if filter == nil {
		log.Printf("get by struct filter failed : %s", errors.WarnInvalidFilter)
		return nil, nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.GetDBWithContext(ctx).
		Where(filter).
		First(ptr)
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
	return ptr, nil
}

func (gx *gormX[T, ID, PT]) GetByMapFilter(ctx context.Context, filter map[string]any) (PT, error) {
	if len(filter) == 0 {
		log.Printf("get by map filter failed : %s", errors.WarnInvalidFilter)
		return nil, nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.GetDBWithContext(ctx).
		Where(filter).
		First(ptr)
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
	return ptr, nil
}

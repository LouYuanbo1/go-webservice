package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
)

func (gx *gormX[T, ID, PT]) GetByID(ctx context.Context, id ID) (PT, error) {
	if IsZero(id) {
		log.Printf("get by id failed : %s", WarnInvalidID)
		return nil, nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.GetDBWithContext(ctx).
		First(ptr, id)
	if result.Error != nil {
		log.Printf("get by id failed. table: %s, error: %v", tableName, result.Error)
		return nil, errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("GetByID[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by id failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return ptr, nil
}

func (gx *gormX[T, ID, PT]) GetByStructFilter(ctx context.Context, filter PT) (PT, error) {
	if filter == nil {
		log.Printf("get by struct filter failed : %s", WarnInvalidFilter)
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
		return nil, errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("GetByStructFilter[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by struct filter failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return ptr, nil
}

func (gx *gormX[T, ID, PT]) GetByMapFilter(ctx context.Context, filter map[string]any) (PT, error) {
	if len(filter) == 0 {
		log.Printf("get by map filter failed : %s", WarnInvalidFilter)
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
		return nil, errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("GetByMapFilter[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by map filter failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return ptr, nil
}

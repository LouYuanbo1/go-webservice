package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
)

func (gx *gormX[T, ID, PT]) DeleteByID(ctx context.Context, id ID) error {
	if IsZero(id) {
		log.Printf("delete by id failed : %s", WarnInvalidID)
		return nil
	}

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.GetDBWithContext(ctx).
		Delete(ptr, id)
	if result.Error != nil {
		log.Printf("delete by id %v failed. table: %s, error: %v", id, tableName, result.Error)
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("DeleteByID[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by id failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) DeleteByIDs(ctx context.Context, ids []ID) error {
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

	var model T
	ptr := PT(&model)
	tableName := ptr.TableName()

	result := gx.GetDBWithContext(ctx).
		Delete(ptr, ids)
	if result.Error != nil {
		log.Printf("delete by ids %v failed. table: %s error: %v", ids, tableName, result.Error)
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("DeleteByIDs[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by ids failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) DeleteByStructFilter(ctx context.Context, filter PT) error {
	if filter == nil {
		log.Printf("delete by struct filter failed : %s", WarnInvalidFilter)
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
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("DeleteByStructFilter[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by struct filter failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return nil
}

func (gx *gormX[T, ID, PT]) DeleteByMapFilter(ctx context.Context, filter map[string]any) error {
	if len(filter) == 0 {
		log.Printf("delete by map filter failed : %s", WarnInvalidFilter)
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
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("DeleteByMapFilter[%s]", tableName),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by map filter failed. table: %s, %s", tableName, WarnNoRowsAffected)
	}
	return nil
}

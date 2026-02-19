package internal

import (
	"context"
	"log"

	"github.com/LouYuanbo1/go-webservice/gormx/errors"
)

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
	if len(updateData) == 0 {
		log.Printf("update by map filter failed : %s", errors.WarnInvalidUpdateData)
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

package gen

import (
	"context"
)

func (g *genSession[T, ID, PT]) Update(ctx context.Context, updateData PT) error {
	/*
		if updateData == nil {
			log.Printf("update failed : %s", gormx.WarnInvalidUpdateData)
			return nil
		}

		tableName := updateData.TableName()

		result := gx.GetDBWithContext(ctx).
			Updates(updateData)
		if result.Error != nil {
			log.Printf("update failed. table: %s, error: %v", tableName, result.Error)
			return errorx.New(
				ErrUpdateFailed,
				"gormx",
				fmt.Sprintf("Update[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("update failed. table: %s, %s", tableName, WarnNoRowsAffected)
		}
	*/
	return g.Session.Update(ctx, updateData)
}

func (g *genSession[T, ID, PT]) UpdateByStructFilter(ctx context.Context, filter PT, updateData PT) error {
	/*
		if updateData == nil {
			log.Printf("update by struct filter failed : %s", gormx.WarnInvalidUpdateData)
			return nil
		}
		if filter == nil {
			log.Printf("update by struct filter failed : %s", gormx.WarnInvalidFilter)
			return nil
		}

		tableName := updateData.TableName()

		result := gx.GetDBWithContext(ctx).
			Where(filter).
			Updates(updateData)
		if result.Error != nil {
			log.Printf("update by struct filter %v failed. table: %s error: %v", filter, tableName, result.Error)
			return errorx.New(
				ErrUpdateFailed,
				"gormx",
				fmt.Sprintf("UpdateByStructFilter[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("update by struct filter failed. table: %s, %s", tableName, WarnNoRowsAffected)
		}
	*/
	return g.Session.UpdatesByStructFilter(ctx, filter, updateData)
}

func (g *genSession[T, ID, PT]) UpdateByMapFilter(ctx context.Context, filter map[string]any, updateData map[string]any) error {
	/*
		if len(updateData) == 0 {
			log.Printf("update by map filter failed : %s", gormx.WarnInvalidUpdateData)
			return nil
		}
		if len(filter) == 0 {
			log.Printf("update by map filter failed : %s", gormx.WarnInvalidFilter)
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
			return errorx.New(
				ErrUpdateFailed,
				"gormx",
				fmt.Sprintf("UpdateByMapFilter[%s]", tableName),
				result.Error,
			)
		}
		if result.RowsAffected == 0 {
			log.Printf("update by map filter failed. table: %s, %s", tableName, WarnNoRowsAffected)
		}
	*/
	var model T
	ptr := PT(&model)
	return g.Session.UpdatesByMapFilter(ctx, ptr, filter, updateData)
}

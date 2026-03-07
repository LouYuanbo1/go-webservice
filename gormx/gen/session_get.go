package gen

import (
	"context"
	"log"

	"github.com/LouYuanbo1/go-webservice/gormx"
)

func (g *genSession[T, ID, PT]) GetByID(ctx context.Context, id ID) (PT, error) {
	if IsZero(id) {
		log.Printf("get by id failed : %s", gormx.WarnInvalidID)
		return nil, nil
	}

	var model T
	ptr := PT(&model)
	/*
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
	*/
	g.Session.GetByID(ctx, ptr, id)
	return ptr, nil
}

func (g *genSession[T, ID, PT]) GetByStructFilter(ctx context.Context, filter PT) (PT, error) {
	/*
		if filter == nil {
			log.Printf("get by struct filter failed : %s", gormx.WarnInvalidFilter)
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
	*/
	var model T
	ptr := PT(&model)
	g.Session.GetByStructFilter(ctx, ptr, filter)
	return ptr, nil
}

func (g *genSession[T, ID, PT]) GetByMapFilter(ctx context.Context, filter map[string]any) (PT, error) {
	/*
		if len(filter) == 0 {
			log.Printf("get by map filter failed : %s", gormx.WarnInvalidFilter)
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
	*/
	var model T
	ptr := PT(&model)
	g.Session.GetByMapFilter(ctx, ptr, filter)
	return ptr, nil
}

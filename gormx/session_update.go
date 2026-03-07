package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
)

func (s *session) Update(ctx context.Context, updateData any) error {
	if updateData == nil {
		log.Printf("update failed : %s", WarnInvalidUpdateData)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Updates(updateData)
	if result.Error != nil {
		log.Printf("update failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrUpdateFailed,
			"gormx",
			fmt.Sprintf("Update[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("update failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (s *session) UpdatesByStructFilter(ctx context.Context, filter any, updateData any) error {
	if updateData == nil {
		log.Printf("updates by struct filter failed : %s", WarnInvalidUpdateData)
		return nil
	}
	if filter == nil {
		log.Printf("updates by struct filter failed : %s", WarnInvalidFilter)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Where(filter).
		Updates(updateData)
	if result.Error != nil {
		log.Printf("updates by struct filter %v failed. table: %s error: %v", filter, result.Statement.Table, result.Error)
		return errorx.New(
			ErrUpdateFailed,
			"gormx",
			fmt.Sprintf("UpdatesByStructFilter[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("updates by struct filter failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (s *session) UpdatesByMapFilter(ctx context.Context, model any, filter map[string]any, updateData map[string]any) error {
	if len(updateData) == 0 {
		log.Printf("updates by map filter failed : %s", WarnInvalidUpdateData)
		return nil
	}
	if len(filter) == 0 {
		log.Printf("updates by map filter failed : %s", WarnInvalidFilter)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Model(model).
		Where(filter).
		Updates(updateData)
	if result.Error != nil {
		log.Printf("updates by map filter %v failed. table: %s error: %v", filter, result.Statement.Table, result.Error)
		return errorx.New(
			ErrUpdateFailed,
			"gormx",
			fmt.Sprintf("UpdatesByMapFilter[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("updates by map filter failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

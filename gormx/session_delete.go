package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
)

func (s *session) DeleteByID(ctx context.Context, model any, id any) error {
	result := s.GetDBWithContext(ctx).
		Delete(model, id)
	if result.Error != nil {
		log.Printf("delete by id %v failed. table: %s, error: %v", id, result.Statement.Table, result.Error)
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("DeleteByID[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by id failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (s *session) DeleteByIDs(ctx context.Context, model any, ids any) error {
	if ids == nil {
		log.Printf("delete by ids failed : %s", WarnEmptyIDSlice)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Delete(model, ids)
	if result.Error != nil {
		log.Printf("delete by ids %v failed. table: %s error: %v", ids, result.Statement.Table, result.Error)
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("DeleteByIDs[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by ids failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (s *session) DeleteByStructFilter(ctx context.Context, model any, filter any) error {
	if filter == nil {
		log.Printf("delete by struct filter failed : %s", WarnInvalidFilter)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Where(filter).
		Delete(model)
	if result.Error != nil {
		log.Printf("delete by struct filter %v failed. table: %s error: %v", filter, result.Statement.Table, result.Error)
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("DeleteByStructFilter[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by struct filter failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (s *session) DeleteByMapFilter(ctx context.Context, model any, filter map[string]any) error {
	if len(filter) == 0 {
		log.Printf("delete by map filter failed : %s", WarnInvalidFilter)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Where(filter).
		Delete(model)
	if result.Error != nil {
		log.Printf("delete by map filter %v failed. table: %s, error: %v", filter, result.Statement.Table, result.Error)
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("DeleteByMapFilter[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by map filter failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

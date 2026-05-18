package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
)

func (s *session) GetByID(ctx context.Context, dest any, id any) error {
	result := s.GetDBWithContext(ctx).
		First(dest, id)
	if result.Error != nil {
		log.Printf("get by id failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("GetByID[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by id failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (s *session) GetByStructFilter(ctx context.Context, dest any, filter any) error {
	if filter == nil {
		log.Printf("get by struct filter failed : %s", WarnInvalidFilter)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Where(filter).
		First(dest)
	if result.Error != nil {
		log.Printf("get by struct filter failed. table: %s, error: %v", result.Statement.Table, result.Error)
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("GetByStructFilter[%s]", result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("get by struct filter failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
)

func (s *session) GetByID(ctx context.Context, dest any, id any) error {
	prefix := "GetByID"
	result := s.GetDBWithContext(ctx).
		First(dest, id)
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (s *session) GetByStructFilter(ctx context.Context, dest any, filter any) error {
	prefix := "GetByStructFilter"
	if filter == nil {
		log.Printf("%s failed: %s", prefix, WarnInvalidFilter)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Where(filter).
		First(dest)
	if result.Error != nil {
		return errorx.New(
			ErrQueryFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("%s failed. table: %s, %s", prefix, result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

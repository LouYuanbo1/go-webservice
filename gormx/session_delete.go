package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
)

func (s *session) DeleteByID(ctx context.Context, model any, id any) error {
	prefix := "DeleteByID"
	result := s.GetDBWithContext(ctx).
		Delete(model, id)
	if result.Error != nil {
		return errorx.New(
			ErrDeleteFailed,
			"gormx",
			fmt.Sprintf("%s[%s]", prefix, result.Statement.Table),
			result.Error,
		)
	}
	if result.RowsAffected == 0 {
		log.Printf("delete by id failed. table: %s, %s", result.Statement.Table, WarnNoRowsAffected)
	}
	return nil
}

func (s *session) DeleteByIDs(ctx context.Context, model any, ids any) error {
	prefix := "DeleteByIDs"
	if ids == nil {
		log.Printf("%s failed : %s", prefix, WarnEmptyIDSlice)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Delete(model, ids)
	if result.Error != nil {
		return errorx.New(
			ErrDeleteFailed,
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

func (s *session) DeleteByStructFilter(ctx context.Context, model any, filter any) error {
	prefix := "DeleteByStructFilter"
	if filter == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidFilter)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Where(filter).
		Delete(model)
	if result.Error != nil {
		return errorx.New(
			ErrDeleteFailed,
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

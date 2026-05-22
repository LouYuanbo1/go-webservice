package gormx

import (
	"context"
	"fmt"
	"log"

	"github.com/LouYuanbo1/go-webservice/errorx"
)

func (s *session) Update(ctx context.Context, updateData any) error {
	prefix := "Update"
	if updateData == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidUpdateData)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Updates(updateData)
	if result.Error != nil {
		return errorx.New(
			ErrUpdateFailed,
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

func (s *session) UpdatesByStructFilter(ctx context.Context, filter any, updateData any) error {
	prefix := "UpdatesByStructFilter"
	if updateData == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidUpdateData)
		return nil
	}
	if filter == nil {
		log.Printf("%s failed : %s", prefix, WarnInvalidFilter)
		return nil
	}

	result := s.GetDBWithContext(ctx).
		Where(filter).
		Updates(updateData)
	if result.Error != nil {
		return errorx.New(
			ErrUpdateFailed,
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

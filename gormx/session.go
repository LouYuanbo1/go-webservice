package gormx

import (
	"context"

	"gorm.io/gorm"
)

// 这里的 Session 就是当前事务的会话，你可以在事务内通过它执行 SQL 语句，而无需关心底层是事务还是普通连接。

type Session interface {
	GetDBWithContext(ctx context.Context) *gorm.DB
	Create(ctx context.Context, model any, opts ...ConflictOption) error
	CreateInBatches(ctx context.Context, models any, batchSize int, opts ...ConflictOption) error
	GetByID(ctx context.Context, dest any, id any) error
	GetByStructFilter(ctx context.Context, dest any, filter any) error
	GetByMapFilter(ctx context.Context, dest any, filter map[string]any) error
	FindByIDs(ctx context.Context, dest any, ids any, opts ...OrderOption) error
	FindByStructFilter(ctx context.Context, dest any, filter any, opts ...OrderOption) error
	FindByMapFilter(ctx context.Context, dest any, filter map[string]any, opts ...OrderOption) error
	FindByPage(ctx context.Context, dest any, primaryKey string, page, pageSize int, opts ...OrderOption) error
	FindByCursor(ctx context.Context, dest any, primaryKey string, cursor any, limit int) error
	FindInBatches(
		ctx context.Context,
		dest any,
		batchSize int,
		callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
		opts ...OrderOption,
	) error
	FindInBatchesByStructFilter(
		ctx context.Context,
		dest any,
		filter any,
		batchSize int,
		callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
		opts ...OrderOption,
	) error
	FindInBatchesByMapFilter(
		ctx context.Context,
		dest any,
		filter map[string]any,
		batchSize int,
		callback func(ctx context.Context, tx *gorm.DB, batch int, models any) error,
		opts ...OrderOption,
	) error
	Update(ctx context.Context, updateData any) error
	UpdatesByStructFilter(ctx context.Context, filter any, updateData any) error
	UpdatesByMapFilter(ctx context.Context, model any, filter map[string]any, updateData map[string]any) error
	DeleteByID(ctx context.Context, model any, id any) error
	DeleteByIDs(ctx context.Context, model any, ids any) error
	DeleteByStructFilter(ctx context.Context, model any, filter any) error
	DeleteByMapFilter(ctx context.Context, model any, filter map[string]any) error
}

type session struct {
	db *gorm.DB
}

func NewSession(db *gorm.DB) Session {
	return &session{db: db}
}

/*
func (s *session) GetDBWithContext(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if !ok {
		return s.db.WithContext(ctx)
	}
	return tx.WithContext(ctx)
}
*/

func (s *session) GetDBWithContext(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx)
}

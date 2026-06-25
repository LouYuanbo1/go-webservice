package gormx

import (
	"context"

	"gorm.io/gorm"
)

// TxExecutor 事务执行器，纯粹SQL数据执行，不包含其他逻辑
type TxExecutor struct {
	exec *Executor
}

func (tx *TxExecutor) Exec(ctx context.Context, fn func(gormDB *gorm.DB) error) error {
	return tx.exec.Exec(ctx, fn)
}

func (tx *TxExecutor) Create[T any, PT PointerModel[T]](ctx context.Context, model PT, opts ...ConflictOption) error {
	return tx.exec.Create(ctx, model, opts...)
}

func (tx *TxExecutor) CreateInBatches[T any, PT PointerModel[T]](ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) error {
	return tx.exec.CreateInBatches(ctx, models, batchSize, opts...)
}

func (tx *TxExecutor) GetByID[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest PT, id ID) error {
	return tx.exec.GetByID(ctx, dest, id)
}

func (tx *TxExecutor) GetByFilter[T any, PT PointerModel[T]](ctx context.Context, dest PT, filter PT) error {
	return tx.exec.GetByFilter(ctx, dest, filter)
}

func (tx *TxExecutor) FindByIDs[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest *[]T, ids []ID, opts ...OrderOption) error {
	return tx.exec.FindByIDs(ctx, dest, ids, opts...)
}

func (tx *TxExecutor) FindByFilter[T any, PT PointerModel[T]](ctx context.Context, dest *[]T, filter PT, opts ...OrderOption) error {
	return tx.exec.FindByFilter(ctx, dest, filter, opts...)
}

func (tx *TxExecutor) FindByPage[T any, PT PointerModel[T]](ctx context.Context, dest *[]T, page, pageSize int, opts ...OrderOption) error {
	return tx.exec.FindByPage(ctx, dest, page, pageSize, opts...)
}

func (tx *TxExecutor) FindByCursor[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest *[]T, cursor ID, limit int) error {
	return tx.exec.FindByCursor(ctx, dest, cursor, limit)
}

func (tx *TxExecutor) FindInBatches[T any, PT PointerModel[T]](
	ctx context.Context,
	batchSize int,
	callback func(ctx context.Context, tx *DB, batch int, models *[]T) error,
) error {
	return tx.exec.FindInBatches(ctx, batchSize, callback)
}

func (tx *TxExecutor) Update[T any, PT PointerModel[T]](ctx context.Context, updateData PT) error {
	return tx.exec.Update(ctx, updateData)
}

func (tx *TxExecutor) UpdatesByFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT, updateData PT) error {
	return tx.exec.UpdatesByFilter(ctx, filter, updateData)
}

func (tx *TxExecutor) DeleteByID[T any, PT PointerModel[T], ID comparable](ctx context.Context, id ID) error {
	return tx.exec.DeleteByID[T](ctx, id)
}

func (tx *TxExecutor) DeleteByIDs[T any, PT PointerModel[T], ID comparable](ctx context.Context, ids ...ID) error {
	return tx.exec.DeleteByIDs[T](ctx, ids...)
}

func (tx *TxExecutor) DeleteByFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT) error {
	return tx.exec.DeleteByFilter(ctx, filter)
}

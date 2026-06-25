package gormx

import (
	"context"

	"gorm.io/gorm"
)

// Tx 事务执行器,将来可能添加其他组件，供DB使用
type Tx struct {
	exec *Executor
}

func (tx *Tx) Exec(ctx context.Context, fn func(gormDB *gorm.DB) error) (err error) {
	ctx, span := startSpan(ctx, "Exec")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.Exec(ctx, fn)
	return
}

func (tx *Tx) Create[T any, PT PointerModel[T]](ctx context.Context, model PT, opts ...ConflictOption) (err error) {
	ctx, span := startSpan(ctx, "Create")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.Create(ctx, model, opts...)
	return
}

func (tx *Tx) CreateInBatches[T any, PT PointerModel[T]](ctx context.Context, models []PT, batchSize int, opts ...ConflictOption) (err error) {
	ctx, span := startSpan(ctx, "CreateInBatches")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.CreateInBatches(ctx, models, batchSize, opts...)
	return
}

func (tx *Tx) GetByID[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest PT, id ID) (err error) {
	ctx, span := startSpan(ctx, "GetByID")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.GetByID(ctx, dest, id)
	return
}

func (tx *Tx) GetByFilter[T any, PT PointerModel[T]](ctx context.Context, dest PT, filter PT) (err error) {
	ctx, span := startSpan(ctx, "GetByFilter")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.GetByFilter(ctx, dest, filter)
	return
}

func (tx *Tx) FindByIDs[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest *[]T, ids []ID, opts ...OrderOption) (err error) {
	ctx, span := startSpan(ctx, "FindByIDs")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.FindByIDs(ctx, dest, ids, opts...)
	return
}

func (tx *Tx) FindByPage[T any, PT PointerModel[T]](ctx context.Context, dest *[]T, page, pageSize int, opts ...OrderOption) (err error) {
	ctx, span := startSpan(ctx, "FindByPage")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.FindByPage(ctx, dest, page, pageSize, opts...)
	return
}

func (tx *Tx) FindByFilter[T any, PT PointerModel[T]](ctx context.Context, dest *[]T, filter PT, opts ...OrderOption) (err error) {
	ctx, span := startSpan(ctx, "FindByFilter")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.FindByFilter(ctx, dest, filter, opts...)
	return
}

func (tx *Tx) FindByCursor[T any, PT PointerModel[T], ID comparable](ctx context.Context, dest *[]T, cursor ID, limit int) (err error) {
	ctx, span := startSpan(ctx, "FindByCursor")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.FindByCursor(ctx, dest, cursor, limit)
	return
}

func (tx *Tx) FindInBatches[T any, PT PointerModel[T]](
	ctx context.Context,
	batchSize int,
	callback func(ctx context.Context, tx *DB, batch int, models *[]T) error,
) (err error) {
	ctx, span := startSpan(ctx, "FindInBatches")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.FindInBatches(ctx, batchSize, callback)
	return
}

func (tx *Tx) Update[T any, PT PointerModel[T]](ctx context.Context, updateData PT) (err error) {
	ctx, span := startSpan(ctx, "Update")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.Update(ctx, updateData)
	return
}

func (tx *Tx) UpdatesByFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT, updateData PT) (err error) {
	ctx, span := startSpan(ctx, "UpdatesByFilter")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.UpdatesByFilter(ctx, filter, updateData)
	return
}

func (tx *Tx) DeleteByID[T any, PT PointerModel[T], ID comparable](ctx context.Context, id ID) (err error) {
	ctx, span := startSpan(ctx, "DeleteByID")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.DeleteByID[T](ctx, id)
	return
}

func (tx *Tx) DeleteByIDs[T any, PT PointerModel[T], ID comparable](ctx context.Context, ids ...ID) (err error) {
	ctx, span := startSpan(ctx, "DeleteByIDs")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.DeleteByIDs[T](ctx, ids...)
	return
}

func (tx *Tx) DeleteByFilter[T any, PT PointerModel[T]](ctx context.Context, filter PT) (err error) {
	ctx, span := startSpan(ctx, "DeleteByFilter")
	defer func() {
		endSpan(span, err)
	}()
	err = tx.exec.DeleteByFilter(ctx, filter)
	return
}

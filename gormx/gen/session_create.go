package gen

import (
	"context"
	"log"

	"github.com/LouYuanbo1/go-webservice/gormx"
)

func (g *genSession[T, ID, PT]) Create(ctx context.Context, model PT, opts ...gormx.ConflictOption) error {
	return g.Session.Create(ctx, model, opts...)
}

func (g *genSession[T, ID, PT]) CreateInBatches(ctx context.Context, models []PT, batchSize int, opts ...gormx.ConflictOption) error {
	if len(models) == 0 {
		// 空切片属于合法操作（0 行插入），静默成功更符合批量操作语义
		log.Printf("skipped create in batches: %s", gormx.WarnEmptyModelsSlice)
		return nil
	}
	return g.Session.CreateInBatches(ctx, models, batchSize, opts...)
}

package elasticsearchx

import (
	"context"

	"github.com/LouYuanbo1/go-webservice/elasticsearchx/model"
	"github.com/LouYuanbo1/go-webservice/elasticsearchx/options"
)

type Elasticsearchx[T any, PT model.PointerDocument[T]] interface {
	CreateIndex(ctx context.Context) error
	GetMapIndexCount(ctx context.Context) (map[string]string, error)
	DeleteIndex(ctx context.Context, index string) error
	IndexDoc(ctx context.Context, doc PT) error
	BulkIndexDocs(ctx context.Context, docs []PT, opts ...options.BulkOption) error
	GetDoc(ctx context.Context, index string, id string) (PT, error)
	FindDocsByPages(ctx context.Context, index string, page, size int) ([]PT, error)
	CountDocs(ctx context.Context, index string) (int64, error)
	UpdateDoc(ctx context.Context, doc PT) error
	DeleteDoc(ctx context.Context, index string, id string) error
	BulkDeleteDocs(ctx context.Context, index string, ids []string, opts ...options.BulkOption) error
}

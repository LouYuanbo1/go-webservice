package elasticsearchx

import "context"

type Elasticsearchx interface {
	CreateIndex(ctx context.Context, index string) error
	DeleteIndex(ctx context.Context, index string) error
}

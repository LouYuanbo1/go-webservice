package internal

import (
	"github.com/LouYuanbo1/go-webservice/elasticsearchx/options"
	"github.com/elastic/go-elasticsearch/v9/esutil"
)

func (e *elasticsearchx[T, PT]) bulkIndexerConfigBuilder(opts ...options.BulkOption) *esutil.BulkIndexerConfig {
	bulk := &options.Bulk{}
	bulk.WithNumWorkers(e.bulkIndexerConfig.NumWorkers).
		WithFlushBytes(e.bulkIndexerConfig.FlushBytes).
		WithFlushInterval(e.bulkIndexerConfig.FlushInterval).
		WithOnError(e.bulkIndexerConfig.OnError).
		WithOnFlushStart(e.bulkIndexerConfig.OnFlushStart).
		WithOnFlushEnd(e.bulkIndexerConfig.OnFlushEnd).
		WithTimeout(e.bulkIndexerConfig.Timeout)
	for _, opt := range opts {
		opt(bulk)
	}
	return bulk.Build()
}

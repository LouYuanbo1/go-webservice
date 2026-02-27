package internal

import (
	"github.com/LouYuanbo1/go-webservice/elasticsearchx/options"
	"github.com/elastic/go-elasticsearch/v9/esutil"
)

func (e *elasticsearchx[T, PT]) bulkIndexerConfigBuilder(opts ...options.BulkOption) *esutil.BulkIndexerConfig {
	b := options.NewBulkByConfig(*e.config.BulkIndexer)
	for _, opt := range opts {
		opt(b)
	}
	bulkIndexerConfig := &esutil.BulkIndexerConfig{
		NumWorkers:    b.GetNumWorkers(),
		FlushBytes:    b.GetFlushBytes(),
		FlushInterval: b.GetFlushInterval(),
		OnError:       b.GetOnError(),
		OnFlushStart:  b.GetOnFlushStart(),
		OnFlushEnd:    b.GetOnFlushEnd(),
		Timeout:       b.GetTimeout(),
	}
	return bulkIndexerConfig
}

package elasticsearchx

import (
	"context"
	"time"
)

type BulkIndexerConfig struct {
	NumWorkers    int                                   `mapstructure:"num_workers"`
	FlushBytes    int                                   `mapstructure:"flush_bytes"`
	FlushInterval time.Duration                         `mapstructure:"flush_interval"`
	OnError       func(context.Context, error)          `mapstructure:"on_error"`
	OnFlushStart  func(context.Context) context.Context `mapstructure:"on_flush_start"`
	OnFlushEnd    func(context.Context)                 `mapstructure:"on_flush_end"`
	Timeout       time.Duration                         `mapstructure:"timeout"`
	Stats         bool                                  `mapstructure:"stats"`
}

type Config struct {
	BulkIndexer *BulkIndexerConfig `mapstructure:"bulk_indexer"`
}

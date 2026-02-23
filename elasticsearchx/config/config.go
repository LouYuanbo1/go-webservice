package config

import (
	"context"
	"time"
)

type BulkIndexerConfig struct {
	NumWorkers    int
	FlushBytes    int
	FlushInterval time.Duration
	OnError       func(context.Context, error)
	OnFlushStart  func(context.Context) context.Context
	OnFlushEnd    func(context.Context)
	Timeout       time.Duration
	Stats          bool
}

type ElasticsearchXConfig struct {
	BulkIndexer *BulkIndexerConfig
}

package elasticsearchx

import (
	"context"
	"time"

	"github.com/elastic/go-elasticsearch/v9/esutil"
)

type bulk struct {
	numWorkers    int
	flushBytes    int
	flushInterval time.Duration
	onError       func(context.Context, error)          // Called for indexer errors.
	onFlushStart  func(context.Context) context.Context // Called when the flush starts.
	onFlushEnd    func(context.Context)                 // Called when the flush ends.
	timeout       time.Duration
}

func newBulkByConfig(cfg *BulkIndexerConfig) *bulk {
	return &bulk{
		numWorkers:    cfg.NumWorkers,
		flushBytes:    cfg.FlushBytes,
		flushInterval: cfg.FlushInterval,
		onError:       cfg.OnError,
		onFlushStart:  cfg.OnFlushStart,
		onFlushEnd:    cfg.OnFlushEnd,
		timeout:       cfg.Timeout,
	}
}

type BulkOption func(*bulk)

func WithNumWorkers(numWorkers int) BulkOption {
	return func(b *bulk) {
		b.numWorkers = numWorkers
	}
}

func WithFlushBytes(flushBytes int) BulkOption {
	return func(b *bulk) {
		b.flushBytes = flushBytes
	}
}

func WithFlushInterval(flushInterval time.Duration) BulkOption {
	return func(b *bulk) {
		b.flushInterval = flushInterval
	}
}

func WithOnError(onError func(context.Context, error)) BulkOption {
	return func(b *bulk) {
		b.onError = onError
	}
}

func WithOnFlushStart(onFlushStart func(context.Context) context.Context) BulkOption {
	return func(b *bulk) {
		b.onFlushStart = onFlushStart
	}
}

func WithOnFlushEnd(onFlushEnd func(context.Context)) BulkOption {
	return func(b *bulk) {
		b.onFlushEnd = onFlushEnd
	}
}

func WithTimeout(timeout time.Duration) BulkOption {
	return func(b *bulk) {
		b.timeout = timeout
	}
}

func (e *elasticsearchx[T, PT]) bulkIndexerConfigBuilder(opts ...BulkOption) *esutil.BulkIndexerConfig {
	b := newBulkByConfig(e.config.BulkIndexer)
	for _, opt := range opts {
		opt(b)
	}
	bulkIndexerConfig := &esutil.BulkIndexerConfig{
		NumWorkers:    b.numWorkers,
		FlushBytes:    b.flushBytes,
		FlushInterval: b.flushInterval,
		OnError:       b.onError,
		OnFlushStart:  b.onFlushStart,
		OnFlushEnd:    b.onFlushEnd,
		Timeout:       b.timeout,
	}
	return bulkIndexerConfig
}
